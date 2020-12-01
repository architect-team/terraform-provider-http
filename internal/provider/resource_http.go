package provider

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// resourceHTTP returns the current version of the
// http_resource and needs to be updated when the schema
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
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"request_headers": {
				Type:      schema.TypeMap,
				Optional:  true,
				Sensitive: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"body": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
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
	bodyStr := []byte(body)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(bodyStr))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	for name, value := range headers {
		req.Header.Set(name, value.(string))
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if (method == http.MethodGet || method == http.MethodDelete) && resp.StatusCode == 404 {
		d.SetId("")
		return nil
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return fmt.Errorf("HTTP request error. Response code: %d. Method: %s. Body: %s", resp.StatusCode, method, string(bytes))
	}

	if method == http.MethodGet || method == http.MethodPost {
		id := string(bytes)
		if id != "" {
			d.SetId(id)
		} else {
			return fmt.Errorf("Get/Post endpoints must return the unique id for the http_resource")
		}
	}

	return nil
}

func resourceHTTPCreate(d *schema.ResourceData, meta interface{}) error {
	err := buildHTTPClient(d, http.MethodPost)
	if err != nil {
		return err
	}

	return nil
}

func resourceHTTPRead(d *schema.ResourceData, meta interface{}) error {
	err := buildHTTPClient(d, http.MethodGet)
	if err != nil {
		return err
	}
	return nil
}

func resourceHTTPDelete(d *schema.ResourceData, meta interface{}) error {
	err := buildHTTPClient(d, http.MethodDelete)
	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}

func resourceHTTPUpdate(d *schema.ResourceData, meta interface{}) error {
	err := buildHTTPClient(d, http.MethodPut)
	if err != nil {
		return err
	}

	return nil
}
