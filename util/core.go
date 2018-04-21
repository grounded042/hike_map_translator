package util

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

// GetURLBody gets and returns the body of the passed in url
func GetURLBody(url, password string) ([]byte, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("Error building GET: %v", err)
	}

	req.SetBasicAuth("", password)

	resp, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("GET error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Status error: %v", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Read body: %v", err)
	}

	return data, nil
}
