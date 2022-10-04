package main

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/nsf/jsondiff"
	"github.com/stretchr/testify/assert"
)

var apiGatewayHost string = "api.instill.tech"

var inputJSONArray string = `
[
	{
		"dog": [
			{
				"bbox": [
					319.07828,
					161.427,
					213.41531,
					309.82706
				],
				"score": 0.92880696
			},
			{
				"bbox": [
					140.29298,
					182.36862,
					166.14177,
					279.05853
				],
				"score": 0.865231
			}
		],
		"image_height": 512,
		"image_name": "dog.jpg",
		"image_width": 640
	},
	{
		"dog": [
			{
				"bbox": [
					319.07828,
					161.427,
					213.41531,
					309.82706
				],
				"score": 0.92880696
			},
			{
				"bbox": [
					140.29298,
					182.36862,
					166.14177,
					279.05853
				],
				"score": 0.865231
			}
		],
		"image_height": 512,
		"image_name": "dog.jpg",
		"image_width": 640
	},
	{
		"dog": [
			{
				"bbox": [
					319.07828,
					161.427,
					213.41531,
					309.82706
				],
				"score": 0.92880696
			},
			{
				"bbox": [
					140.29298,
					182.36862,
					166.14177,
					279.05853
				],
				"score": 0.865231
			}
		],
		"image_height": 512,
		"image_name": "dog.jpg",
		"image_width": 640
	}
]`

var inputJSONObject string = `
{
	"self": "https://localhost:8444/resource-servers",
	"kind": "Collection",
	"contents": [
		{
			"self": "https://localhost:8444/resource-servers/rs_N2R3YWoQsDQziHUeUCLXqvucdLJjU8Co",
			"kind": "ResourceServer",
			"id": "rs_N2R3YWoQsDQziHUeUCLXqvucdLJjU8Co",
			"identifier": "instill.tech/inference",
			"name": "Instill Inference API",
			"description": "Instill Inference API is to perform Vision AI inference tasks.",
			"base_url": "https://api.instill.tech",
			"scopes": [
				{
					"name": "infer:classification",
					"description": "Perform classification inference task"
				},
				{
					"name": "infer:detection",
					"description": "Perform detection inference task"
				}
			],
			"signing_algorithm": "HS512",
			"token_lifetime": 86400
		},
		{
			"self": "https://localhost:8444/resource-servers/rs_DOdnaWAD9qTY1ZPLRdoSeX6f0NGe1daN",
			"kind": "ResourceServer",
			"id": "rs_DOdnaWAD9qTY1ZPLRdoSeX6f0NGe1daN",
			"identifier": "instill.tech/management",
			"name": "Instill Management API",
			"description": "Instill Management API is to create or manage your clients programmatically.",
			"base_url": "https://api.instill.tech",
			"scopes": [
				{
					"name": "create:client_grants",
					"description": "Create client grants"
				},
				{
					"name": "create:clients",
					"description": "Create clients"
				}
			],
			"signing_algorithm": "HS512",
			"token_lifetime": 86400
		}
	]
}`

func TestReplaceSelfHost(t *testing.T) {
	{
		outputJSONObjectExpected := `
		{
			"self": "https://api.instill.tech/resource-servers",
			"kind": "Collection",
			"contents": [
				{
					"self": "https://api.instill.tech/resource-servers/rs_N2R3YWoQsDQziHUeUCLXqvucdLJjU8Co",
					"kind": "ResourceServer",
					"id": "rs_N2R3YWoQsDQziHUeUCLXqvucdLJjU8Co",
					"identifier": "instill.tech/inference",
					"name": "Instill Inference API",
					"description": "Instill Inference API is to perform Vision AI inference tasks.",
					"base_url": "https://api.instill.tech",
					"scopes": [
						{
							"name": "infer:classification",
							"description": "Perform classification inference task"
						},
						{
							"name": "infer:detection",
							"description": "Perform detection inference task"
						}
					],
					"signing_algorithm": "HS512",
					"token_lifetime": 86400
				},
				{
					"self": "https://api.instill.tech/resource-servers/rs_DOdnaWAD9qTY1ZPLRdoSeX6f0NGe1daN",
					"kind": "ResourceServer",
					"id": "rs_DOdnaWAD9qTY1ZPLRdoSeX6f0NGe1daN",
					"identifier": "instill.tech/management",
					"name": "Instill Management API",
					"description": "Instill Management API is to create or manage your clients programmatically.",
					"base_url": "https://api.instill.tech",
					"scopes": [
						{
							"name": "create:client_grants",
							"description": "Create client grants"
						},
						{
							"name": "create:clients",
							"description": "Create clients"
						}
					],
					"signing_algorithm": "HS512",
					"token_lifetime": 86400
				}
			]
		}`

		var data interface{}
		err := json.Unmarshal([]byte(inputJSONObject), &data)
		if err != nil {
			t.Error(err.Error())
		}

		// Recursively replace all self values with req.Host
		err = replaceSelfHost(&data, apiGatewayHost)
		if err != nil {
			t.Error(err.Error())
			return
		}

		output, err := json.Marshal(data)
		if err != nil {
			t.Error(err.Error())
		}

		opt := jsondiff.DefaultJSONOptions()

		diff, _ := jsondiff.Compare([]byte(outputJSONObjectExpected), []byte(output), &opt)
		assert.Equal(t, jsondiff.FullMatch, diff, "Expected JSON output is wrong")
	}

	{
		var data interface{}
		err := json.Unmarshal([]byte(inputJSONArray), &data)
		if err != nil {
			t.Error(err.Error())
		}

		// Recursively replace all self values with req.Host
		err = replaceSelfHost(&data, apiGatewayHost)
		if err != nil {
			t.Error(err.Error())
			return
		}

		output, err := json.Marshal(data)
		if err != nil {
			t.Error(err.Error())
		}

		opt := jsondiff.DefaultJSONOptions()
		diff, _ := jsondiff.Compare([]byte(inputJSONArray), []byte(output), &opt)
		assert.Equal(t, jsondiff.FullMatch, diff, "Expected JSON output is wrong")
	}
}

func TestInsertDuration(t *testing.T) {
	{
		var outputJSONArrayExpected string = fmt.Sprintf(`
		[
			{
				"dog": [
					{
						"bbox": [
							319.07828,
							161.427,
							213.41531,
							309.82706
						],
						"score": 0.92880696
					},
					{
						"bbox": [
							140.29298,
							182.36862,
							166.14177,
							279.05853
						],
						"score": 0.865231
					}
				],
				"image_height": 512,
				"image_name": "dog.jpg",
				"image_width": 640,
				"duration": %.3f
			},
			{
				"dog": [
					{
						"bbox": [
							319.07828,
							161.427,
							213.41531,
							309.82706
						],
						"score": 0.92880696
					},
					{
						"bbox": [
							140.29298,
							182.36862,
							166.14177,
							279.05853
						],
						"score": 0.865231
					}
				],
				"image_height": 512,
				"image_name": "dog.jpg",
				"image_width": 640,
				"duration": %.3f
			},
			{
				"dog": [
					{
						"bbox": [
							319.07828,
							161.427,
							213.41531,
							309.82706
						],
						"score": 0.92880696
					},
					{
						"bbox": [
							140.29298,
							182.36862,
							166.14177,
							279.05853
						],
						"score": 0.865231
					}
				],
				"image_height": 512,
				"image_name": "dog.jpg",
				"image_width": 640,
				"duration": %.3f
			}
		]`, time.Millisecond.Seconds(), time.Millisecond.Seconds(), time.Millisecond.Seconds())

		var data interface{}
		err := json.Unmarshal([]byte(inputJSONArray), &data)
		if err != nil {
			t.Error(err.Error())
		}

		// Recursively replace all self values with req.Host
		err = insertDuration(&data, time.Millisecond)
		if err != nil {
			t.Error(err.Error())
			return
		}

		output, err := json.Marshal(data)
		if err != nil {
			t.Error(err.Error())
		}

		opt := jsondiff.DefaultJSONOptions()

		diff, _ := jsondiff.Compare([]byte(outputJSONArrayExpected), []byte(output), &opt)
		assert.Equal(t, jsondiff.FullMatch, diff, "Expected JSON output is wrong")

	}

	{
		var outputJSONObjectExpected string = fmt.Sprintf(`
		{
			"self": "https://localhost:8444/resource-servers",
			"kind": "Collection",
			"duration": %.3f,
			"contents": [
				{
					"self": "https://localhost:8444/resource-servers/rs_N2R3YWoQsDQziHUeUCLXqvucdLJjU8Co",
					"kind": "ResourceServer",
					"id": "rs_N2R3YWoQsDQziHUeUCLXqvucdLJjU8Co",
					"identifier": "instill.tech/inference",
					"name": "Instill Inference API",
					"description": "Instill Inference API is to perform Vision AI inference tasks.",
					"base_url": "https://api.instill.tech",
					"scopes": [
						{
							"name": "infer:classification",
							"description": "Perform classification inference task"
						},
						{
							"name": "infer:detection",
							"description": "Perform detection inference task"
						}
					],
					"signing_algorithm": "HS512",
					"token_lifetime": 86400
				},
				{
					"self": "https://localhost:8444/resource-servers/rs_DOdnaWAD9qTY1ZPLRdoSeX6f0NGe1daN",
					"kind": "ResourceServer",
					"id": "rs_DOdnaWAD9qTY1ZPLRdoSeX6f0NGe1daN",
					"identifier": "instill.tech/management",
					"name": "Instill Management API",
					"description": "Instill Management API is to create or manage your clients programmatically.",
					"base_url": "https://api.instill.tech",
					"scopes": [
						{
							"name": "create:client_grants",
							"description": "Create client grants"
						},
						{
							"name": "create:clients",
							"description": "Create clients"
						}
					],
					"signing_algorithm": "HS512",
					"token_lifetime": 86400
				}
			]
		}`, time.Millisecond.Seconds())

		var data interface{}
		err := json.Unmarshal([]byte(inputJSONObject), &data)
		if err != nil {
			t.Error(err.Error())
		}

		// Recursively replace all self values with req.Host
		err = insertDuration(&data, time.Millisecond)
		if err != nil {
			t.Error(err.Error())
			return
		}

		output, err := json.Marshal(data)
		if err != nil {
			t.Error(err.Error())
		}

		opt := jsondiff.DefaultJSONOptions()

		diff, _ := jsondiff.Compare([]byte(outputJSONObjectExpected), []byte(output), &opt)
		assert.Equal(t, jsondiff.FullMatch, diff, "Expected JSON output is wrong")
	}
}
