package hcplog

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/pkg/errors"
	"github.com/gobwas/glob"
)

// Config represents Config paramets for the hcplog tool
type Config struct {
	Account        string
	LandscapeHost  string
	AccessEndpoint string
	ConfigEndpoint string
	User           string
	Password       string
}

// FileInfo maintains information about a log file in the HANA Cloud Platform
type FileInfo struct {
	Name         string `json:"name"`
	Size         uint64 `json:"size"`
	Description  string `json:"description"`
	LastModified uint64 `json:"lastModified"`
	ProcessID    string `json:"processId"`
}

type fileInfosPayload struct {
	LogFiles []FileInfo `json:"logFiles"`
}

// Client is a Log client for the HANA Cloud Platform which is scoped to the
// configured account and offers operations to list log files for an app, grab
// (download) log files for an app and set log levels for an app
type Client struct {
	Config *Config
	http   *http.Client
}

// ModifiedTime encapsulates a time
type ModifiedTime struct {
	time.Time
}

var _ fmt.Stringer = &Config{}

func (c *Config) String() string {
	return fmt.Sprintf("(Account=%s,LandscapeHost=%s,User=%s,Password=%s)", c.Account, c.LandscapeHost, c.User, strings.Repeat("*", len(c.Password)))
}

func (f *FileInfo) String() string {
	return fmt.Sprintf("(Name=%s,Size=%s,Description=%s,ProcessID=%s,LastModified=%d)",
		f.Name, f.Size, f.Description, f.ProcessID, f.LastModified)
}

// NewClient creates and initializes a log client for given configuraation
func NewClient(config Config) (*Client, error) {
	_, err := url.Parse(config.LandscapeHost)
	if err != nil {
		return nil, errors.Wrapf(err, "Landscape host: '%s' does not appear valid", config.LandscapeHost)
	}
	config.AccessEndpoint = "https://logapi." + config.LandscapeHost + "/log/api_basic/v1/logs/"
	config.ConfigEndpoint = "https://logconfig." + config.LandscapeHost + "/log/api_basic/v1/logs/"

	return &Client{
		Config: &config,
		http: &http.Client{
			Timeout: 4 * time.Minute,
		},
	}, nil
}

// ListFiles lists the log files for the given application
func (c *Client) ListFiles(app string) ([]FileInfo, error) {
	logsEndpoint := c.Config.AccessEndpoint + c.Config.Account + "/" + app + "/web"
	log.Println("GET ", logsEndpoint, "...")
	req, err := http.NewRequest("GET", logsEndpoint, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create http request for URL %s", logsEndpoint)
	}
	req.SetBasicAuth(c.Config.User, c.Config.Password)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to list files for %s:%s", c.Config.Account, app)
	}
	if resp.StatusCode != 200 {
		return nil, errors.Errorf("Failed to read logs for %s:%s due to %s", c.Config.Account, app, resp.Status)
	}
	defer resp.Body.Close()
	// body, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	return nil, errors.Wrapf(err, "Failed to read log file list for: %s:%s", c.Config.Account, app)
	// }
	// fmt.Printf("Body is \n %s \n", string(body))
	return ParseLogList(resp.Body)
}

func (c *Client) Download(app string, filename string) error {
	logsEndpoint := c.Config.AccessEndpoint + c.Config.Account + "/" + app + "/web/" + filename
	log.Println("GET ", logsEndpoint, "...")
	req, err := http.NewRequest("GET", logsEndpoint, nil)
	if err != nil {
		return  errors.Wrapf(err, "Failed to create http request for URL %s", logsEndpoint)
	}
	req.SetBasicAuth(c.Config.User, c.Config.Password)
	resp, err := c.http.Do(req)
	if err != nil {
		return errors.Wrapf(err, "Failed to download file %s of %s:%s", filename, c.Config.Account, app)
	}
	if resp.StatusCode != 200 {
		return errors.Errorf("Failed to download file %s of %s:%s due to %s", filename, c.Config.Account, app, resp.Status)
	}
	defer resp.Body.Close()
	dir, err := os.Getwd()
	if err != nil {
		return errors.Wrapf(err, "Failed to download file %s of %s:%s", filename, c.Config.Account, app)
	}
	fpath := filepath.Join(dir, filename)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrapf(err, "Failed to download file %s of %s:%s", filename, c.Config.Account, app)
	}
	err = ioutil.WriteFile(fpath, body, 0644)
	if err != nil {
		return errors.Wrapf(err, "Failed to write file %s of %s:%s", fpath, c.Config.Account, app)
	}
	fmt.Printf("Downloaded %s\n", fpath)
	return nil
}

// PrintFiles print the log file list for the given application on the given writer in a tabular form
func (c *Client) PrintFiles(app string, w io.Writer) error {
	fileInfos, err := c.ListFiles(app)
	if err != nil {
		return err
	}
	tw := tabwriter.NewWriter(w, 6, 4, 2, ' ', 0)
	defer tw.Flush()
	fmt.Fprintf(tw, "\n %s\t%s\t%s\t%s\t%s", "Name", "Description", "Size", "ProcessID", "LastModified")
	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].LastModified < fileInfos[j].LastModified
	})
	for i := range fileInfos {
		f := &fileInfos[i]
		fmt.Fprintf(tw, "\n %s\t%s\t%d\t%s\t%d", f.Name, f.Description, f.Size, f.ProcessID, f.LastModified)
	}
	// fmt.Println(fileInfos)
	return nil
}


func (c *Client) GrabFilesAndPrint(app string, names []string ,w io.Writer) error {
	fileInfos, err := c.ListFiles(app)
	if err != nil {
		return  err
	}
	for _, name := range names {
		for _, fi := range fileInfos {
			g := glob.MustCompile(name)
			if g.Match(fi.Name) {
				fmt.Printf("%s matches %s\n", fi.Name, name)
				c.Download(app, fi.Name)
			}
		}
	}
	return nil
}



// ParseLogList parses JSON input from the given reader into a slice of log file infos.
func ParseLogList(r io.Reader) ([]FileInfo, error) {
	var payload fileInfosPayload
	err := json.NewDecoder(r).Decode(&payload)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to parse log file list")
	}
	return payload.LogFiles, nil
}
