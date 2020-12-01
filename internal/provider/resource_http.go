package provider

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// resourceHTTP returns the current version of the
// http resource and needs to be updated when the schema
// version is incremented.
func resourceHTTP() *schema.Resource { return resourceHTTPV1() }

func resourceHTTPV1() *schema.Resource {
	return &schema.Resource{
		Create: resourceHTTPCreate,
		Read:   resourceHTTPRead,
		Delete: resourceHTTPDelete,
		Update: resourceHTTPUpdate,
		// MigrateState:  resourceHTTPMigrateState,
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"url": {
				Type:     schema.TypeString,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"request_headers": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"body": {
				Type:     schema.TypeString,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func buildHTTPClient(d *schema.ResourceData, method string) error {
	url := d.Get("url").(string)
	headers := d.Get("request_headers").(map[string]interface{})
	body := d.Get("body").(string)
	reader := strings.NewReader(body)

	client := &http.Client{}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return err
	}

	for name, value := range headers {
		req.Header.Set(name, value.(string))
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP request error. Response code: %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" || isContentTypeText(contentType) == false {
		return fmt.Errorf("Content-Type is not recognized as a text type, got %q", contentType)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	responseHeaders := make(map[string]string)
	for k, v := range resp.Header {
		// Concatenate according to RFC2616
		// cf. https://www.w3.org/Protocols/rfc2616/rfc2616-sec4.html#sec4.2
		responseHeaders[k] = strings.Join(v, ", ")
	}

	/*
		d.Set("body", string(bytes))

		// set ID as something more stable than time
		d.SetId(url)
	*/
	if method == "Get" || method == "Post" {
		id := string(bytes)
		if id != "" {
			d.SetId(string(bytes))
		} else {
			return fmt.Errorf("Get/Post endpoints must return the unique id for the http_resource")
		}
	}

	return nil
}

func resourceHTTPCreate(d *schema.ResourceData, meta interface{}) error {
	err := buildHTTPClient(d, "Post")
	if err != nil {
		return err
	}

	return nil
}

func resourceHTTPRead(d *schema.ResourceData, meta interface{}) error {
	err := buildHTTPClient(d, "Get")
	if err != nil {
		return err
	}
	return nil
}

func resourceHTTPDelete(d *schema.ResourceData, meta interface{}) error {
	err := buildHTTPClient(d, "Delete")
	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}

func resourceHTTPUpdate(d *schema.ResourceData, meta interface{}) error {
	err := buildHTTPClient(d, "Put")
	if err != nil {
		return err
	}

	return nil
}
