import http from "k6/http";
import { check, group } from "k6";

import { connectorHost } from "./const.js"
import { deepEqual } from "./helper.js"

import {
    genAuthHeader,
} from "../helpers.js";

export function CheckList(data) {

    group("Connector API: List source connector definitions", () => {

        check(http.request("GET", `${connectorHost}/v1alpha/source-connector-definitions`, null, { headers: genAuthHeader(data.userAccessToken, "application/json") }), {
            "GET /v1alpha/source-connector-definitions response status is 200": (r) => r.status === 200,
            "GET /v1alpha/source-connector-definitions response has source_connector_definitions array": (r) => Array.isArray(r.json().source_connector_definitions),
            "GET /v1alpha/source-connector-definitions response total_size > 0": (r) => r.json().total_size > 0
        });

        var limitedRecords = http.request("GET", `${connectorHost}/v1alpha/source-connector-definitions`, null, { headers: genAuthHeader(data.userAccessToken, "application/json") })
        check(http.request("GET", `${connectorHost}/v1alpha/source-connector-definitions?page_size=0`), {
            "GET /v1alpha/source-connector-definitions?page_size=0 response status is 200": (r) => r.status === 200,
            "GET /v1alpha/source-connector-definitions?page_size=0 response all records": (r) => r.json().source_connector_definitions.length === limitedRecords.json().source_connector_definitions.length,
        });

        check(http.request("GET", `${connectorHost}/v1alpha/source-connector-definitions?page_size=1`, null, { headers: genAuthHeader(data.userAccessToken, "application/json") }), {
            "GET /v1alpha/source-connector-definitions?page_size=1 response status is 200": (r) => r.status === 200,
            "GET /v1alpha/source-connector-definitions?page_size=1 response source_connector_definitions size 1": (r) => r.json().source_connector_definitions.length === 1,
        });

        var pageRes = http.request("GET", `${connectorHost}/v1alpha/source-connector-definitions?page_size=1`, null, { headers: genAuthHeader(data.userAccessToken, "application/json") })
        check(http.request("GET", `${connectorHost}/v1alpha/source-connector-definitions?page_size=1&page_token=${pageRes.json().next_page_token}`), {
            [`GET /v1alpha/source-connector-definitions?page_size=1&page_token=${pageRes.json().next_page_token} response status is 200`]: (r) => r.status === 200,
            [`GET /v1alpha/source-connector-definitions?page_size=1&page_token=${pageRes.json().next_page_token} response source_connector_definitions size 1`]: (r) => r.json().source_connector_definitions.length === 1,
        });

        check(http.request("GET", `${connectorHost}/v1alpha/source-connector-definitions?page_size=1&view=VIEW_BASIC`, null, { headers: genAuthHeader(data.userAccessToken, "application/json") }), {
            "GET /v1alpha/source-connector-definitions?page_size=1&view=VIEW_BASIC response status 200": (r) => r.status === 200,
            "GET /v1alpha/source-connector-definitions?page_size=1&view=VIEW_BASIC response source_connector_definitions[0].connector_definition.spec is null": (r) => r.json().source_connector_definitions[0].connector_definition.spec === null,
        });

        check(http.request("GET", `${connectorHost}/v1alpha/source-connector-definitions?page_size=1&view=VIEW_FULL`, null, { headers: genAuthHeader(data.userAccessToken, "application/json") }), {
            "GET /v1alpha/source-connector-definitions?page_size=1&view=VIEW_FULL response status 200": (r) => r.status === 200,
            "GET /v1alpha/source-connector-definitions?page_size=1&view=VIEW_FULL response source_connector_definitions[0].connector_definition.spec is not null": (r) => r.json().source_connector_definitions[0].connector_definition.spec !== null,
        });

        check(http.request("GET", `${connectorHost}/v1alpha/source-connector-definitions?page_size=1`, null, { headers: genAuthHeader(data.userAccessToken, "application/json") }), {
            "GET /v1alpha/source-connector-definitions?page_size=1 response status 200": (r) => r.status === 200,
            "GET /v1alpha/source-connector-definitions?page_size=1 response source_connector_definitions[0].connector_definition.spec is null": (r) => r.json().source_connector_definitions[0].connector_definition.spec === null,
        });

        check(http.request("GET", `${connectorHost}/v1alpha/source-connector-definitions?page_size=${limitedRecords.json().total_size}`, null, { headers: genAuthHeader(data.userAccessToken, "application/json") }), {
            [`GET /v1alpha/source-connector-definitions?page_size=${limitedRecords.json().total_size} response status 200`]: (r) => r.status === 200,
            [`GET /v1alpha/source-connector-definitions?page_size=${limitedRecords.json().total_size} response next_page_token is empty`]: (r) => r.json().next_page_token === "",
        });
    });
}

export function CheckGet(data) {
    group("Connector API: Get source connector definition", () => {
        var allRes = http.request("GET", `${connectorHost}/v1alpha/source-connector-definitions`, null, { headers: genAuthHeader(data.userAccessToken, "application/json") })
        var def = allRes.json().source_connector_definitions[0]
        check(http.request("GET", `${connectorHost}/v1alpha/source-connector-definitions/${def.id}`, null, { headers: genAuthHeader(data.userAccessToken, "application/json") }), {
            [`GET /v1alpha/source-connector-definitions/${def.id} response status is 200`]: (r) => r.status === 200,
            [`GET /v1alpha/source-connector-definitions/${def.id} response has the exact record`]: (r) => deepEqual(r.json().source_connector_definition, def),
            [`GET /v1alpha/source-connector-definitions/${def.id} response has the non-empty resource name ${def.name}`]: (r) => r.json().source_connector_definition.name != "",
            [`GET /v1alpha/source-connector-definitions/${def.id} response has the resource name ${def.name}`]: (r) => r.json().source_connector_definition.name === def.name,
        });

        check(http.request("GET", `${connectorHost}/v1alpha/source-connector-definitions/${def.id}?view=VIEW_BASIC`, null, { headers: genAuthHeader(data.userAccessToken, "application/json") }), {
            [`GET /v1alpha/source-connector-definitions/${def.id}?view=VIEW_BASIC response status 200`]: (r) => r.status === 200,
            [`GET /v1alpha/source-connector-definitions/${def.id}?view=VIEW_BASIC response source_connector_definition.connector_definition.spec is null`]: (r) => r.json().source_connector_definition.connector_definition.spec === null,
        });

        check(http.request("GET", `${connectorHost}/v1alpha/source-connector-definitions/${def.id}?view=VIEW_FULL`, null, { headers: genAuthHeader(data.userAccessToken, "application/json") }), {
            [`GET /v1alpha/source-connector-definitions/${def.id}?view=VIEW_FULL response status 200`]: (r) => r.status === 200,
            [`GET /v1alpha/source-connector-definitions/${def.id}?view=VIEW_FULL response source_connector_definition.connector_definition.spec is not null`]: (r) => r.json().source_connector_definition.connector_definition.spec !== null,
        });

        check(http.request("GET", `${connectorHost}/v1alpha/source-connector-definitions/${def.id}`, null, { headers: genAuthHeader(data.userAccessToken, "application/json") }), {
            [`GET /v1alpha/source-connector-definitions/${def.id} response status 200`]: (r) => r.status === 200,
            [`GET /v1alpha/source-connector-definitions/${def.id} response source_connector_definition.connector_definition.spec is null`]: (r) => r.json().source_connector_definition.connector_definition.spec === null,
        });
    });
}
