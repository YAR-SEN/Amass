// Copyright 2017 Jeff Foley. All rights reserved.
// Use of this source code is governed by Apache 2 LICENSE that can be found in the LICENSE file.

package amass

import (
	"encoding/json"
	"net/url"

	"github.com/OWASP/Amass/amass/utils"
)

// CertDB is the AmassService that handles access to the CertDB data source.
type CertDB struct {
	BaseAmassService

	SourceType string
}

// NewCertDB returns he object initialized, but not yet started.
func NewCertDB(e *Enumeration) *CertDB {
	c := &CertDB{SourceType: CERT}

	c.BaseAmassService = *NewBaseAmassService(e, "CertDB", c)
	return c
}

// OnStart implements the AmassService interface
func (c *CertDB) OnStart() error {
	c.BaseAmassService.OnStart()

	go c.startRootDomains()
	return nil
}

// OnStop implements the AmassService interface
func (c *CertDB) OnStop() error {
	c.BaseAmassService.OnStop()
	return nil
}

func (c *CertDB) startRootDomains() {
	// Look at each domain provided by the config
	for _, domain := range c.Enum().Config.Domains() {
		c.executeQuery(domain)
	}
}

func (c *CertDB) executeQuery(domain string) {
	u := c.getURL(domain)
	page, err := utils.RequestWebPage(u, nil, nil, "", "")
	if err != nil {
		c.Enum().Log.Printf("%s: %s: %v", c.String(), u, err)
		return
	}

	var names []string
	if err := json.Unmarshal([]byte(page), &names); err != nil {
		c.Enum().Log.Printf("%s: Failed to unmarshal JSON: %v", c.String(), err)
		return
	}

	c.SetActive()
	re := c.Enum().Config.DomainRegex(domain)
	for _, name := range names {
		n := re.FindString(name)
		if n == "" {
			continue
		}

		req := &AmassRequest{
			Name:   cleanName(n),
			Domain: domain,
			Tag:    c.SourceType,
			Source: c.String(),
		}

		if c.Enum().DupDataSourceName(req) {
			continue
		}
		c.Enum().Bus.Publish(NEWNAME, req)
	}
}

func (c *CertDB) getURL(domain string) string {
	u, _ := url.Parse("https://certdb.com/api")

	u.RawQuery = url.Values{
		"q":             {domain},
		"response_type": {"3"},
	}.Encode()
	return u.String()
}