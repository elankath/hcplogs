package hcplog

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
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
	LogFiles []*FileInfo `json:"logFiles"`
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
			Timeout: time.Minute,
		},
	}, nil
}

// ListFiles lists the log files for the given application
func (c *Client) ListFiles(app string) ([]FileInfo, error) {
	logsEndpoint := c.Config.AccessEndpoint + c.Config.Account + "/" + app + "/web"
	log.Println("GET ", logsEndpoint)
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
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to read log file list for: %s:%s", c.Config.Account, app)
	}
	fmt.Printf("body is \n %s \n", string(body))
	defer resp.Body.Close()
	return nil, nil
}

// PrintFiles print the log file list for the given appliation to the given writer
func (c *Client) PrintFiles(app string, w io.Writer) error {
	_, err := c.ListFiles(app)
	return err
}

// ParseLogList parses JSON input from the given reader into a slice of log file infos.
func ParseLogList(r io.Reader) (fileInfos []*FileInfo, err error) {
	var payload *fileInfosPayload
	if err = json.NewDecoder(r).Decode(&payload); err != nil {
		fileInfos = payload.LogFiles
	}
	return
}

func main() {
	fmt.Println("vim-go")
}
