import http from "k6/http";
import { check, group } from "k6";
import {
  randomString
} from "https://jslib.k6.io/k6-utils/1.1.0/index.js";

import { connectorHost } from "./connector-backend/const.js"

import * as sourceConnectorDefinition from './connector-backend/rest-source-connector-definition.js';
import * as destinationConnectorDefinition from './connector-backend/rest-destination-connector-definition.js';
import * as sourceConnector from './connector-backend/rest-source-connector.js';
import * as destinationConnector from './connector-backend/rest-destination-connector.js';

import {
  genAuthHeader,
  generateUserToken,
  deleteUser,
  deleteUserClients,
} from "./helpers.js";

export let options = {
  setupTimeout: '300s',
  insecureSkipTLSVerify: true,
  thresholds: {
    checks: ["rate == 1.0"],
  },
};

const testUserEmail = `test_${randomString(10)}@foo.bar`;

const hydraHost = "https://127.0.0.1:4445";
const kratosHost = "https://127.0.0.1:4434";
const apiHost = "https://127.0.0.1:8000"; //

export function setup() {
  const [testUser, userAccessToken] = generateUserToken(
    kratosHost,
    hydraHost,
    apiHost,
    testUserEmail
  );
  return {
    user: testUser,
    userAccessToken: userAccessToken,
  };
}

export default function (data) {

  /*
   * Connector API - API CALLS
   */

  // Health check
  group("Connector API: Health check", () => {
    check(http.request("GET", `${connectorHost}/v1alpha/health/connector`), {
      "GET /v1alpha/health/connector response status is 200": (r) => r.status === 200,
    });
  });

  // Source connector definitions
  sourceConnectorDefinition.CheckList(data)
  sourceConnectorDefinition.CheckGet(data)

  // Destination connector definitions
  destinationConnectorDefinition.CheckList(data)
  destinationConnectorDefinition.CheckGet(data)

  // Source connectors
  sourceConnector.CheckCreate(data)
  sourceConnector.CheckList(data)
  sourceConnector.CheckGet(data)
  sourceConnector.CheckUpdate(data)
  sourceConnector.CheckDelete(data)
  sourceConnector.CheckLookUp(data)
  sourceConnector.CheckState(data)
  sourceConnector.CheckRename(data)

  // Destination connectors
  destinationConnector.CheckCreate(data)
  destinationConnector.CheckList(data)
  destinationConnector.CheckGet(data)
  destinationConnector.CheckUpdate(data)
  destinationConnector.CheckLookUp(data)
  destinationConnector.CheckState(data)
  destinationConnector.CheckRename(data)
  destinationConnector.CheckWrite(data)

}

export function teardown(data) {
  
  group("Connector API: Delete all source connector created by this test", () => {
    for (const srcConnector of http
      .request("GET", `${connectorHost}/v1alpha/source-connectors`, null, { headers: genAuthHeader(data.userAccessToken, "application/json") })
      .json("source_connectors")) {
      check(
        http.request("DELETE", `${connectorHost}/v1alpha/source-connectors/${srcConnector.id}`, null, { headers: genAuthHeader(data.userAccessToken, "application/json") }),
        {
          [`DELETE /v1alpha/source-connectors/${srcConnector.id} response status is 204`]: (r) =>
            r.status === 204,
        }
      );
    }
  });

  group("Connector API: Delete all destination connector created by this test", () => {
    for (const desConnector of http
      .request("GET", `${connectorHost}/v1alpha/destination-connectors`, null, { headers: genAuthHeader(data.userAccessToken, "application/json") })
      .json("destination_connectors")) {
      check(
        http.request("DELETE", `${connectorHost}/v1alpha/destination-connectors/${desConnector.id}`, null, { headers: genAuthHeader(data.userAccessToken, "application/json") }),
        {
          [`DELETE /v1alpha/destination-connectors/${desConnector.id} response status is 204`]: (r) =>
            r.status === 204,
        }
      );
    }
  });  

  deleteUserClients(hydraHost, data.user.id);
  deleteUser(kratosHost, data.user.id);
}
