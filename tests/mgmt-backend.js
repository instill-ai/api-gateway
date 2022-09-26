import encoding from "k6/encoding";
import http from "k6/http";
import {check, group, sleep} from "k6";
import {randomString} from "https://jslib.k6.io/k6-utils/1.1.0/index.js";

import {
  generateUserToken,
  generateClientToken,
  deleteUser,
  deleteUserClients,
  genAuthHeader,
  genNoAuthHeader,
} from "./helpers.js";

const hydraHost = "https://127.0.0.1:4445";
const kratosHost = "https://127.0.0.1:4434";
const apiHost = "https://127.0.0.1:8000"; // set as api gateway url, the same as HYDRA_AUDIENCE in .env.dev

const testCuboUserEmail = `test_${randomString(10)}@yunyun.cloud`;
const testNonCuboUserEmail = `test_${randomString(10)}@foo.bar`;

const cuboModelID = "m_3v2Yq6ocICEq0LxDdt8dBtl92Yl3QeWA";
const cuboModelVersion = 1;

const allowedAudiences = [
  "instill.tech/inference",
  "instill.tech/management",
  apiHost,
];

export let options = {
  insecureSkipTLSVerify: true,
  thresholds: {
    checks: ["rate == 1.0"],
  },
};

export function setup() {
  // Import cubo test user and generate a cubo user token
  const [testCuboUser, cuboUserAccessToken] = generateUserToken(
    kratosHost,
    hydraHost,
    apiHost,
    testCuboUserEmail
  );

  // Create a client for cubo test user and generate a cubo client token
  const [testCuboClient, cuboClientAccessToken] = generateClientToken(
    cuboUserAccessToken,
    apiHost,
    apiHost
  );

  // Import non-cubo test user and generate a non-cubo user token
  const [testNonCuboUser, nonCuboUserAccessToken] = generateUserToken(
    kratosHost,
    hydraHost,
    apiHost,
    testNonCuboUserEmail
  );

  // Create a client for non-cubo test user and generate a non-cubo client token
  const [testNonCuboClient, nonCuboClientAccessToken] = generateClientToken(
    nonCuboUserAccessToken,
    apiHost,
    apiHost
  );

  return {
    cubo: {
      user: testCuboUser,
      client: testCuboClient,
      userAccessToken: cuboUserAccessToken,
      clientAccessToken: cuboClientAccessToken,
    },
    nonCubo: {
      user: testNonCuboUser,
      client: testNonCuboClient,
      userAccessToken: nonCuboUserAccessToken,
      clientAccessToken: nonCuboClientAccessToken,
    },
  };
}

export default function (data) {
  var client, clientName, clientDesc;
  const allowdAuthHeaders = {
    "User-token": genAuthHeader(
      data.nonCubo.userAccessToken,
      "application/json"
    ),
    "Client-token": genAuthHeader(
      data.nonCubo.clientAccessToken,
      "application/json"
    ),
  };

  /*
   * Management API - API CALLS
   */

  // Health check
  {
    group("Management API: Health check", () => {
      check(http.request("GET", `${apiHost}/health/management`), {
        "GET /management/health response status is 200": (r) =>
          r.status === 200,
      });
    });
  }

  // mgmt-backend consumes user token or client token, here we
  // test both kinds of tokens
  for (const tokenType in allowdAuthHeaders) {
    const authHeader = allowdAuthHeaders[tokenType];
    // Clients
    {
      group(`Management API [with ${tokenType}]: Create a client`, () => {
        clientName = "Test Client";
        clientDesc = "This is a test client.";
        client = http.request(
          "POST",
          `${apiHost}/clients`,
          JSON.stringify({
            name: clientName,
            description: clientDesc,
          }),
          {
            headers: authHeader,
          }
        );

        check(client, {
          "POST /clients response status is 201": (r) => r.status === 201,
          "POST /clients response id check": (r) => r.json("id") !== undefined,
          "POST /clients response secret check": (r) =>
            r.json("secret") !== undefined,
          "POST /clients response name check": (r) =>
            r.json("name") === clientName,
          "POST /clients response description check": (r) =>
            r.json("description") === clientDesc,
          "POST /clients response created_at check": (r) =>
            r.json("created_at") !== undefined,
          "POST /clients response updated_at check": (r) =>
            r.json("updated_at") !== undefined,
          "POST /clients response kind check": (r) =>
            r.json("kind") === "Client",
          "POST /clients response self check": (r) =>
            r.json("self") === `${apiHost}/clients/${r.json("id")}`,
          "POST /clients response duration": (r) =>
            r.json("duration") !== undefined,
        });
        check(
          http.request("POST", `${apiHost}/clients`, null, {
            headers: authHeader,
          }),
          {
            "POST /clients response status is 400": (r) => r.status === 400,
          }
        );
        check(
          http.request(
            "POST",
            `${apiHost}/clients`,
            JSON.stringify({
              name: clientName,
              description: clientDesc,
            }),
            {
              headers: genNoAuthHeader("application/json"),
            }
          ),
          {
            "POST /clients response status is 401": (r) => r.status === 401,
          }
        );
      });

      group(`Management API: [with ${tokenType}] Get a client`, () => {
        check(
          http.request("GET", `${apiHost}/clients/${client.json("id")}`, null, {
            headers: authHeader,
          }),
          {
            [`GET /clients/${client.json("id")} response status is 200`]: (r) =>
              r.status === 200,
            [`GET /clients/${client.json("id")} response id check`]: (r) =>
              r.json("id") === client.json("id"),
            [`GET /clients/${client.json("id")} response secret check`]: (r) =>
              r.json("secret") === undefined,
            [`GET /clients/${client.json("id")} response name check`]: (r) =>
              r.json("name") === clientName,
            [`GET /clients/${client.json("id")} response description check`]: (
              r
            ) => r.json("description") === clientDesc,
            [`GET /clients/${client.json("id")} response created_at check`]: (
              r
            ) => r.json("created_at") === client.json("created_at"),
            [`GET /clients/${client.json("id")} response updated_at check`]: (
              r
            ) => r.json("updated_at") !== undefined,
            [`GET /clients/${client.json("id")} response kind check`]: (r) =>
              r.json("kind") === "Client",
            [`GET /clients/${client.json("id")} response self check`]: (r) =>
              r.json("self") === `${apiHost}/clients/${r.json("id")}`,
            [`GET /clients/${client.json("id")} response duration`]: (r) =>
              r.json("duration") !== undefined,
          }
        );
        check(
          http.request("GET", `${apiHost}/clients/non_exist_client_id`, null, {
            headers: authHeader,
          }),
          {
            "GET /clients/non_exist_client_id response status is 404": (r) =>
              r.status === 404,
          }
        );
        check(
          http.request("GET", `${apiHost}/clients/${client.json("id")}`, null, {
            headers: genNoAuthHeader("application/json"),
          }),
          {
            [`GET /clients/${client.json("id")} response status is 401`]: (r) =>
              r.status === 401,
          }
        );
      });

      group(`Management API: [with ${tokenType}] Update a client`, () => {
        clientName = "Updated test client";
        clientDesc = "This is an updated test client.";
        let updatedClient = http.request(
          "PATCH",
          `${apiHost}/clients/${client.json("id")}`,
          JSON.stringify({
            name: clientName,
            description: clientDesc,
          }),
          {
            headers: authHeader,
          }
        );

        check(updatedClient, {
          [`PATCH /clients/${client.json("id")} response status is 200`]: (r) =>
            r.status === 200,
          [`PATCH /clients/${client.json("id")} response id check`]: (r) =>
            r.json("id") === client.json("id"),
          [`PATCH /clients/${client.json("id")} response secret check`]: (r) =>
            r.json("secret") === undefined,
          [`PATCH /clients/${client.json("id")} response name check`]: (r) =>
            r.json("name") === clientName,
          [`PATCH /clients/${client.json("id")} response description check`]: (
            r
          ) => r.json("description") === clientDesc,
          [`PATCH /clients response created_at check`]: (r) =>
            r.json("created_at") === client.json("created_at"),
          [`PATCH /clients response updated_at changed check`]: (r) =>
            r.json("updated_at") !== client.json("updated_at"),
          [`PATCH /clients/${client.json("id")} response kind check`]: (r) =>
            r.json("kind") === "Client",
          [`PATCH /clients/${client.json("id")} response self check`]: (r) =>
            r.json("self") === `${apiHost}/clients/${r.json("id")}`,
          [`PATCH /clients/${client.json("id")} response duration`]: (r) =>
            r.json("duration") !== undefined,
        });

        // update the client for subsequent tests
        client = updatedClient;

        check(
          http.request(
            "PATCH",
            `${apiHost}/clients/${client.json("id")}`,
            null,
            {
              headers: authHeader,
            }
          ),
          {
            [`PATCH /clients/${client.json("id")} response status is 400`]: (
              r
            ) => r.status === 400,
          }
        );
        check(
          http.request(
            "PATCH",
            `${apiHost}/clients/non_exist_client_id`,
            JSON.stringify({
              name: clientName,
              description: clientDesc,
            }),
            {
              headers: authHeader,
            }
          ),
          {
            "PATCH /clients/non_exist_client_id response status is 404": (r) =>
              r.status === 404,
          }
        );
        check(
          http.request(
            "PATCH",
            `${apiHost}/clients/${client.json("id")}`,
            JSON.stringify({
              name: clientName,
              description: clientDesc,
            }),
            {
              headers: genNoAuthHeader("application/json"),
            }
          ),
          {
            [`PATCH /clients/${client.json("id")} response status is 401`]: (
              r
            ) => r.status === 401,
          }
        );
      });

      group(
        `Management API: [with ${tokenType}] Rotate a client secret`,
        () => {
          let rotatedClient = http.request(
            "POST",
            `${apiHost}/clients/${client.json("id")}/rotate-secret`,
            null,
            {
              headers: authHeader,
            }
          );
          check(rotatedClient, {
            [`POST /clients/${client.json(
              "id"
            )}/rotate-secret response status is 200`]: (r) => r.status === 200,
            [`POST /clients/${client.json(
              "id"
            )}/rotate-secret response id check`]: (r) =>
              r.json("id") === client.json("id"),
            [`POST /clients/${client.json(
              "id"
            )}/rotate-secret response secret check`]: (r) =>
              r.json("secret") !== undefined,
            [`POST /clients/${client.json(
              "id"
            )}/rotate-secret response secret rotated check`]: (r) =>
              r.json("secret") !== client.json("secret"),
            [`POST /clients/${client.json(
              "id"
            )}/rotate-secret response name check`]: (r) =>
              r.json("name") === clientName,
            [`POST /clients/${client.json(
              "id"
            )}/rotate-secret response description check`]: (r) =>
              r.json("description") === clientDesc,
            [`POST /clients/${client.json(
              "id"
            )}/rotate-secret response created_at check`]: (r) =>
              r.json("created_at") === client.json("created_at"),
            [`POST /clients/${client.json(
              "id"
            )}/rotate-secret response updated_at check`]: (r) =>
              r.json("updated_at") !== undefined,
            [`POST /clients/${client.json(
              "id"
            )}/rotate-secret response kind check`]: (r) =>
              r.json("kind") === "Client",
            [`POST /clients/${client.json(
              "id"
            )}/rotate-secret response self check`]: (r) =>
              r.json("self") === `${apiHost}/clients/${r.json("id")}`,
            [`POST /clients/${client.json(
              "id"
            )}/rotate-secret response duration`]: (r) =>
              r.json("duration") !== undefined,
          });

          // update the client for subsequent tests
          client = rotatedClient;

          check(
            http.request(
              "POST",
              `${apiHost}/clients/non_exist_client_id/rotate-secret`,
              null,
              {
                headers: authHeader,
              }
            ),
            {
              "POST /clients/non_exist_client_id/rotate-secret response status is 404":
                (r) => r.status === 404,
            }
          );
          check(
            http.request(
              "POST",
              `${apiHost}/clients/${client.json("id")}/rotate-secret`,
              JSON.stringify({
                name: clientName,
                description: clientDesc,
              }),
              {
                headers: genNoAuthHeader("application/json"),
              }
            ),
            {
              [`POST /clients/${client.json(
                "id"
              )}/rotate-secret response status is 401`]: (r) =>
                r.status === 401,
            }
          );
        }
      );

      group(`Management API: [with ${tokenType}] Get clients`, () => {
        check(
          http.request("GET", `${apiHost}/clients`, null, {
            headers: authHeader,
          }),
          {
            "GET /clients response status is 200": (r) => r.status === 200,
            "GET /clients response contents check": (r) =>
              r.json("contents").length >= 0,
            "GET /clients response contents[-1] id": (r) =>
              r.json("contents").slice(-1)[0].id === client.json("id"),
            "GET /clients response contents[0] secret": (r) =>
              r.json("contents").slice(-1)[0].secret === undefined,
            "GET /clients response contents[0] name": (r) =>
              r.json("contents").slice(-1)[0].name === client.json("name"),
            "GET /clients response contents[0] description": (r) =>
              r.json("contents").slice(-1)[0].description ===
              client.json("description"),
            "GET /clients response contents[0] created_at": (r) =>
              r.json("contents").slice(-1)[0].created_at !== undefined,
            "GET /clients response contents[0] updated_at": (r) =>
              r.json("contents").slice(-1)[0].updated_at !== undefined,
            "GET /clients response kind check": (r) =>
              r.json("kind") === "Clients",
            "GET /clients response self check": (r) =>
              r.json("self") === `${apiHost}/clients`,
            "GET /clients response duration": (r) =>
              r.json("duration") !== undefined,
          }
        );
      });
    }
    // Delete the client
    {
      group(`Management API: [with ${tokenType}] Delete a client`, () => {
        check(
          http.request(
            "DELETE",
            `${apiHost}/clients/${client.json("id")}`,
            null,
            {
              headers: authHeader,
            }
          ),
          {
            [`DELETE /clients/${client.json("id")} response status is 204`]: (
              r
            ) => r.status === 204,
          }
        );
        check(
          http.request(
            "DELETE",
            `${apiHost}/clients/non_exist_client_id`,
            null,
            {
              headers: authHeader,
            }
          ),
          {
            "DELETE /clients/non_exist_client_id response status is 404": (r) =>
              r.status === 404,
          }
        );
        check(
          http.request("DELETE", `${apiHost}/clients/${client.json("id")}`, {
            headers: genNoAuthHeader("application/json"),
          }),
          {
            [`DELETE /clients/${client.json("id")} response status is 401`]: (
              r
            ) => r.status === 401,
          }
        );
      });
    }
  }

  // Tokens
  {
    // Get client Token
    group(`Management API: Get a client token`, () => {
      let basicAuthHeader = encoding.b64encode(
        `${data.nonCubo.client.id}:${data.nonCubo.client.secret}`
      );

      for (const aud of allowedAudiences) {
        let res = http.request(
          "POST",
          `${apiHost}/oauth2/token?audience=${aud}&grant_type=client_credentials`,
          null,
          {
            headers: {
              "Content-Type": "application/json",
              Authorization: `Basic ${basicAuthHeader}`,
            },
          }
        );

        check(res, {
          [`POST /oauth2/token?audience=${aud}&grant_type=client_credentials response status is 200`]:
            (r) => r.status === 200,
          [`POST /oauth2/token?audience=${aud}&grant_type=client_credentials response kind`]:
            (r) => r.json("kind") === "OAuth2JWT",
          [`POST /oauth2/token?audience=${aud}&grant_type=client_credentials response access_token issued successfully)}`]:
            (r) => r.json("access_token") !== undefined,
          [`POST /oauth2/token?audience=${aud}&grant_type=client_credentials response expires_in`]:
            (r) => r.json("expires_in") !== undefined,
          [`POST /oauth2/token?audience=${aud}&grant_type=client_credentials response expiry`]:
            (r) => r.json("expiry") !== undefined,
          [`POST /oauth2/token?audience=${aud}&grant_type=client_credentials response token_type`]:
            (r) => r.json("token_type") === "Bearer",
          [`POST /oauth2/token?audience=${aud}&grant_type=client_credentials response duration`]:
            (r) => r.json("duration") !== undefined,
        });

        check(
          http.request(
            "POST",
            `${apiHost}/oauth2/token?audience=${aud}`,
            null,
            {
              headers: {
                "Content-Type": "application/json",
                Authorization: `Basic ${basicAuthHeader}`,
              },
            }
          ),
          {
            [`POST /oauth2/token?audience=${aud} response status is 400`]: (
              r
            ) => r.status === 400,
          }
        );

        check(
          http.request(
            "POST",
            `${apiHost}/oauth2/token?audience=${aud}&grant_type=non_supported_grant_type`,
            null,
            {
              headers: {
                "Content-Type": "application/json",
                Authorization: `Basic ${basicAuthHeader}`,
              },
            }
          ),
          {
            [`POST /oauth2/token?audience=${aud}&grant_type=non_supported_grant_type response status is 400`]:
              (r) => r.status === 400,
          }
        );

        check(
          http.request(
            "POST",
            `${apiHost}/oauth2/token?audience=${aud}&grant_type=client_credentials`,
            null,
            {
              headers: genNoAuthHeader("application/json"),
            }
          ),
          {
            [`POST /oauth2/token?audience=${aud}&grant_type=client_credentials response status is 401\n`]:
              (r) => r.status === 401,
          }
        );
      }

      check(
        http.request("POST", `${apiHost}/oauth2/token`, null, {
          headers: {
            "Content-Type": "application/json",
            Authorization: `Basic ${basicAuthHeader}`,
          },
        }),
        {
          [`POST /oauth2/token response status is 400`]: (r) =>
            r.status === 400,
        }
      );
      check(
        http.request(
          "POST",
          `${apiHost}/oauth2/token?audience=non_supported_audience&grant_type=client_credentials`,
          null,
          {
            headers: {
              "Content-Type": "application/json",
              Authorization: `Basic ${basicAuthHeader}`,
            },
          }
        ),
        {
          [`POST /oauth2/token?audience=non_supported_audience&grant_type=client_credentials response status is 400`]:
            (r) => r.status === 400,
        }
      );
    });

    // Get cubo client Token for accessing cubo models
    group(`Management API: Get a cubo client token for cubo model`, () => {
      let basicAuthHeader = encoding.b64encode(
        `${data.cubo.client.id}:${data.cubo.client.secret}`
      );

      for (const aud of allowedAudiences) {
        let res = http.request(
          "POST",
          `${apiHost}/oauth2/token?audience=${aud}&grant_type=client_credentials`,
          JSON.stringify({
            contents: [
              {
                model_id: cuboModelID,
                task: "detection",
                version: cuboModelVersion,
              },
            ],
          }),
          {
            headers: {
              "Content-Type": "application/json",
              Authorization: `Basic ${basicAuthHeader}`,
            },
          }
        );

        check(res, {
          [`POST /oauth2/token?audience=${aud}&grant_type=client_credentials response status is 200`]:
            (r) => r.status === 200,
          [`POST /oauth2/token?audience=${aud}&grant_type=client_credentials response kind`]:
            (r) => r.json("kind") === "OAuth2JWT",
          [`POST /oauth2/token?audience=${aud}&grant_type=client_credentials response access_token issued successfully)}`]:
            (r) => r.json("access_token") !== undefined,
          [`POST /oauth2/token?audience=${aud}&grant_type=client_credentials response expires_in`]:
            (r) => r.json("expires_in") !== undefined,
          [`POST /oauth2/token?audience=${aud}&grant_type=client_credentials response expiry`]:
            (r) => r.json("expiry") !== undefined,
          [`POST /oauth2/token?audience=${aud}&grant_type=client_credentials response token_type`]:
            (r) => r.json("token_type") === "Bearer",
          [`POST /oauth2/token?audience=${aud}&grant_type=client_credentials response duration`]:
            (r) => r.json("duration") !== undefined,
        });

        check(
          http.request(
            "POST",
            `${apiHost}/oauth2/token?audience=${aud}&grant_type=client_credentials`,
            JSON.stringify({
              contents: [
                {
                  model_id: "invalid_model",
                  task: "detection",
                  version: cuboModelVersion,
                },
              ],
            }),
            {
              headers: {
                "Content-Type": "application/json",
                Authorization: `Basic ${basicAuthHeader}`,
              },
            }
          ),
          {
            [`POST /oauth2/token?audience=${aud}&grant_type=client_credentials response status is 400`]:
              (r) => r.status === 400,
          }
        );

        check(
          http.request(
            "POST",
            `${apiHost}/oauth2/token?audience=${aud}&grant_type=client_credentials`,
            JSON.stringify({
              contents: [
                {
                  model_id: cuboModelID,
                  task: "invalid_task",
                  version: cuboModelVersion,
                },
              ],
            }),
            {
              headers: {
                "Content-Type": "application/json",
                Authorization: `Basic ${basicAuthHeader}`,
              },
            }
          ),
          {
            [`POST /oauth2/token?audience=${aud}&grant_type=client_credentials response status is 400`]:
              (r) => r.status === 400,
          }
        );

        check(
          http.request(
            "POST",
            `${apiHost}/oauth2/token?audience=${aud}&grant_type=client_credentials`,
            JSON.stringify({
              contents: [
                {
                  model_id: cuboModelID,
                  task: "detection",
                  version: 2,
                },
              ],
            }),
            {
              headers: {
                "Content-Type": "application/json",
                Authorization: `Basic ${basicAuthHeader}`,
              },
            }
          ),
          {
            [`POST /oauth2/token?audience=${aud}&grant_type=client_credentials response status is 400`]:
              (r) => r.status === 400,
          }
        );

        check(
          http.request(
            "POST",
            `${apiHost}/oauth2/token?audience=${aud}&grant_type=client_credentials`,
            null,
            {
              headers: genNoAuthHeader("application/json"),
            }
          ),
          {
            [`POST /oauth2/token?audience=${aud}&grant_type=client_credentials response status is 401\n`]:
              (r) => r.status === 401,
          }
        );
      }

      check(
        http.request(
          "POST",
          `${apiHost}/oauth2/token`,
          JSON.stringify({
            contents: [
              {
                model_id: cuboModelID,
                task: "detection",
                version: cuboModelVersion,
              },
            ],
          }),
          {
            headers: {
              "Content-Type": "application/json",
              Authorization: `Basic ${basicAuthHeader}`,
            },
          }
        ),
        {
          [`POST /oauth2/token response status is 400`]: (r) =>
            r.status === 400,
        }
      );
      check(
        http.request(
          "POST",
          `${apiHost}/oauth2/token?audience=non_supported_audience&grant_type=client_credentials`,
          JSON.stringify({
            contents: [
              {
                model_id: cuboModelID,
                task: "detection",
                version: cuboModelVersion,
              },
            ],
          }),
          {
            headers: {
              "Content-Type": "application/json",
              Authorization: `Basic ${basicAuthHeader}`,
            },
          }
        ),
        {
          [`POST /oauth2/token?audience=non_supported_audience&grant_type=client_credentials response status is 400`]:
            (r) => r.status === 400,
        }
      );
    });

    // Get non-cubo client Token for accessing cubo models
    group(`Management API: Get a non-cubo client token for cubo model`, () => {
      let basicAuthHeader = encoding.b64encode(
        `${data.nonCubo.client.id}:${data.nonCubo.client.secret}`
      );

      for (const aud of allowedAudiences) {
        check(
          http.request(
            "POST",
            `${apiHost}/oauth2/token?audience=${aud}&grant_type=client_credentials`,
            JSON.stringify({
              contents: [
                {
                  model_id: cuboModelID,
                  task: "detection",
                  version: cuboModelVersion,
                },
              ],
            }),
            {
              headers: {
                "Content-Type": "application/json",
                Authorization: `Basic ${basicAuthHeader}`,
              },
            }
          ),
          {
            [`POST /oauth2/token?audience=${aud}&grant_type=client_credentials response status is 403`]:
              (r) => r.status === 403,
          }
        );

        check(
          http.request(
            "POST",
            `${apiHost}/oauth2/token?audience=${aud}&grant_type=client_credentials`,
            JSON.stringify({
              contents: [
                {
                  model_id: "invalid_model",
                  task: "detection",
                  version: cuboModelVersion,
                },
              ],
            }),
            {
              headers: {
                "Content-Type": "application/json",
                Authorization: `Basic ${basicAuthHeader}`,
              },
            }
          ),
          {
            [`POST /oauth2/token?audience=${aud}&grant_type=client_credentials response status is 403`]:
              (r) => r.status === 403,
          }
        );

        check(
          http.request(
            "POST",
            `${apiHost}/oauth2/token?audience=${aud}&grant_type=client_credentials`,
            JSON.stringify({
              contents: [
                {
                  model_id: cuboModelID,
                  task: "invalid_task",
                  version: cuboModelVersion,
                },
              ],
            }),
            {
              headers: {
                "Content-Type": "application/json",
                Authorization: `Basic ${basicAuthHeader}`,
              },
            }
          ),
          {
            [`POST /oauth2/token?audience=${aud}&grant_type=client_credentials response status is 400`]:
              (r) => r.status === 400,
          }
        );

        check(
          http.request(
            "POST",
            `${apiHost}/oauth2/token?audience=${aud}&grant_type=client_credentials`,
            JSON.stringify({
              contents: [
                {
                  model_id: cuboModelID,
                  task: "detection",
                  version: 2,
                },
              ],
            }),
            {
              headers: {
                "Content-Type": "application/json",
                Authorization: `Basic ${basicAuthHeader}`,
              },
            }
          ),
          {
            [`POST /oauth2/token?audience=${aud}&grant_type=client_credentials response status is 403`]:
              (r) => r.status === 403,
          }
        );

        check(
          http.request(
            "POST",
            `${apiHost}/oauth2/token?audience=${aud}&grant_type=client_credentials`,
            null,
            {
              headers: genNoAuthHeader("application/json"),
            }
          ),
          {
            [`POST /oauth2/token?audience=${aud}&grant_type=client_credentials response status is 401\n`]:
              (r) => r.status === 401,
          }
        );
      }

      check(
        http.request(
          "POST",
          `${apiHost}/oauth2/token`,
          JSON.stringify({
            contents: [
              {
                model_id: cuboModelID,
                task: "detection",
                version: cuboModelVersion,
              },
            ],
          }),
          {
            headers: {
              "Content-Type": "application/json",
              Authorization: `Basic ${basicAuthHeader}`,
            },
          }
        ),
        {
          [`POST /oauth2/token response status is 400`]: (r) =>
            r.status === 400,
        }
      );
      check(
        http.request(
          "POST",
          `${apiHost}/oauth2/token?audience=non_supported_audience&grant_type=client_credentials`,
          JSON.stringify({
            contents: [
              {
                model_id: cuboModelID,
                task: "detection",
                version: cuboModelVersion,
              },
            ],
          }),
          {
            headers: {
              "Content-Type": "application/json",
              Authorization: `Basic ${basicAuthHeader}`,
            },
          }
        ),
        {
          [`POST /oauth2/token?audience=non_supported_audience&grant_type=client_credentials response status is 400`]:
            (r) => r.status === 400,
        }
      );
    });
  }
  sleep(1);
}

export function teardown(data) {
  deleteUserClients(hydraHost, data.cubo.user.id);
  deleteUser(kratosHost, data.cubo.user.id);
  deleteUserClients(hydraHost, data.nonCubo.user.id);
  deleteUser(kratosHost, data.nonCubo.user.id);
}
