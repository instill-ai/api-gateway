package main

import (
	"fmt"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestURLRegexp_FindStringSubmatch(t *testing.T) {
	c := qt.New(t)

	const namespace, contentID = "summer-wombat", "llava-34b"
	repository := fmt.Sprintf("%s/%s", namespace, contentID)
	urlPrefix := fmt.Sprintf("/v2/%s/", repository)

	testcases := []struct {
		path     string
		mismatch bool

		resourceType string
		resourceID   string
	}{
		{
			path:     "/v2/",
			mismatch: true,
		},
		{
			path:         urlPrefix + "manifests/latest",
			resourceType: "manifests",
			resourceID:   "latest",
		},
		{
			path:         urlPrefix + "blobs/uploads",
			resourceType: "blobs",
			resourceID:   "uploads",
		},
		{
			path:         urlPrefix + "blobs/uploads/a48adcae-2a9e-4929-8379-5b92a6a68821",
			resourceType: "blobs",
			resourceID:   "uploads/a48adcae-2a9e-4929-8379-5b92a6a68821",
		},
	}

	for _, tc := range testcases {
		c.Run(tc.path, func(c *qt.C) {
			matches := urlRegexp.FindStringSubmatch(tc.path)
			if tc.mismatch {
				c.Check(matches, qt.HasLen, 0)
				return
			}
			c.Assert(len(matches) > 0, qt.IsTrue)
			c.Check(matches[1], qt.Equals, repository)
			c.Check(matches[2], qt.Equals, namespace)
			c.Check(matches[3], qt.Equals, contentID)
			c.Check(matches[4], qt.Equals, tc.resourceType)
			c.Check(matches[5], qt.Equals, tc.resourceID)
		})
	}

}
