import http from "k6/http";
import { check, group } from "k6";
import { randomString } from "https://jslib.k6.io/k6-utils/1.1.0/index.js";

import { connectorHost, modelHost, pipelineHost } from "./const.js"

import * as constant from "./const.js"
import * as helper from "./helper.js"

import {
    genAuthHeader,
} from "../helpers.js";

export function CheckCreate(data) {

    group("Connector API: Create source connectors", () => {

        var httpSrcConnector = {
            "id": "source-http",
            "source_connector_definition": constant.httpSrcDefRscName,
            "connector": {
                "description": "HTTP source",
                "configuration": {},
            }
        }

        var gRPCSrcConnector = {
            "id": "source-grpc",
            "source_connector_definition": constant.gRPCSrcDefRscName,
            "connector": {
                "description": "gRPC source",
                "configuration": {},
            }
        }

        var resSrcHTTP = http.request(
            "POST",
            `${connectorHost}/v1alpha/source-connectors`,
            JSON.stringify(httpSrcConnector), {
            headers: genAuthHeader(data.userAccessToken, "application/json"),
        })

        check(resSrcHTTP, {
            "POST /v1alpha/source-connectors response status for creating HTTP source connector 201": (r) => r.status === 201,
            "POST /v1alpha/source-connectors response connector name": (r) => r.json().source_connector.name == `source-connectors/${httpSrcConnector.id}`,
            "POST /v1alpha/source-connectors response connector uid": (r) => helper.isUUID(r.json().source_connector.uid),
            "POST /v1alpha/source-connectors response connector source_connector_definition": (r) => r.json().source_connector.source_connector_definition === constant.httpSrcDefRscName
        });

        check(http.request(
            "POST",
            `${connectorHost}/v1alpha/source-connectors`,
            JSON.stringify(httpSrcConnector), {
            headers: genAuthHeader(data.userAccessToken, "application/json"),
        }), {
            "POST /v1alpha/source-connectors response duplicate HTTP source connector status 409": (r) => r.status === 409
        });

        var resSrcGRPC = http.request(
            "POST",
            `${connectorHost}/v1alpha/source-connectors`,
            JSON.stringify(gRPCSrcConnector), {
            headers: genAuthHeader(data.userAccessToken, "application/json"),
        })

        check(resSrcGRPC, {
            "POST /v1alpha/source-connectors response status for creating gRPC source connector 201": (r) => r.status === 201,
        });

        check(http.request(
            "POST",
            `${connectorHost}/v1alpha/source-connectors`,
            {}, {
            headers: genAuthHeader(data.userAccessToken, "application/json"),
        }), {
            "POST /v1alpha/source-connectors response status for creating empty body 400": (r) => r.status === 400,
        });

        // Delete test records
        check(http.request("DELETE", `${connectorHost}/v1alpha/source-connectors/${resSrcHTTP.json().source_connector.id}`, null, { headers: genAuthHeader(data.userAccessToken, "application/json") }), {
            [`DELETE /v1alpha/source-connectors/${resSrcHTTP.json().source_connector.id} response status 204`]: (r) => r.status === 204,
        });
        check(http.request("DELETE", `${connectorHost}/v1alpha/source-connectors/${resSrcGRPC.json().source_connector.id}`, null, { headers: genAuthHeader(data.userAccessToken, "application/json") }), {
            [`DELETE /v1alpha/source-connectors/${resSrcGRPC.json().source_connector.id} response status 204`]: (r) => r.status === 204,
        });
    });
}

export function CheckList(data) {

    group("Connector API: List source connectors", () => {

        check(http.request("GET", `${connectorHost}/v1alpha/source-connectors`, null, { headers: genAuthHeader(data.userAccessToken, "application/json") }), {
            [`GET /v1alpha/source-connectors response status is 200`]: (r) => r.status === 200,
            [`GET /v1alpha/source-connectors response source_connectors array is 0 length`]: (r) => r.json().source_connectors.length === 0,
            [`GET /v1alpha/source-connectors response next_page_token is empty`]: (r) => r.json().next_page_token === "",
            [`GET /v1alpha/source-connectors response total_size is 0`]: (r) => r.json().next_page_token == 0,
        });

        var reqBodies = [];
        reqBodies[0] = {
            "id": "source-http",
            "source_connector_definition": constant.httpSrcDefRscName,
            "connector": {
                "configuration": {}
            }
        }

        reqBodies[1] = {
            "id": "source-grpc",
            "source_connector_definition": constant.gRPCSrcDefRscName,
            "connector": {
                "configuration": {}
            }
        }

        // Create connectors
        for (const reqBody of reqBodies) {
            check(http.request(
                "POST",
                `${connectorHost}/v1alpha/source-connectors`,
                JSON.stringify(reqBody), {
                headers: genAuthHeader(data.userAccessToken, "application/json"),
            }), {
                [`POST /v1alpha/source-connectors x${reqBodies.length} response status 201`]: (r) => r.status === 201,
            });
        }

        check(http.request("GET", `${connectorHost}/v1alpha/source-connectors`, null, { headers: genAuthHeader(data.userAccessToken, "application/json") }), {
            [`GET /v1alpha/source-connectors response status is 200`]: (r) => r.status === 200,
            [`GET /v1alpha/source-connectors response has source_connectors array`]: (r) => Array.isArray(r.json().source_connectors),
            [`GET /v1alpha/source-connectors response has total_size = ${reqBodies.length}`]: (r) => r.json().total_size == reqBodies.length,
        });

        var limitedRecords = http.request("GET", `${connectorHost}/v1alpha/source-connectors`, null, { headers: genAuthHeader(data.userAccessToken, "application/json") })
        check(http.request("GET", `${connectorHost}/v1alpha/source-connectors?page_size=0`), {
            "GET /v1alpha/source-connectors?page_size=0 response status is 200": (r) => r.status === 200,
            "GET /v1alpha/source-connectors?page_size=0 response all records": (r) => r.json().source_connectors.length === limitedRecords.json().source_connectors.length,
        });

        check(http.request("GET", `${connectorHost}/v1alpha/source-connectors?page_size=1`, null, { headers: genAuthHeader(data.userAccessToken, "application/json") }), {
            "GET /v1alpha/source-connectors?page_size=1 response status is 200": (r) => r.status === 200,
            "GET /v1alpha/source-connectors?page_size=1 response source_connectors size 1": (r) => r.json().source_connectors.length === 1,
        });

        var pageRes = http.request("GET", `${connectorHost}/v1alpha/source-connectors?page_size=1`, null, { headers: genAuthHeader(data.userAccessToken, "application/json") })
        check(http.request("GET", `${connectorHost}/v1alpha/source-connectors?page_size=1&page_token=${pageRes.json().next_page_token}`), {
            [`GET /v1alpha/source-connectors?page_size=1&page_token=${pageRes.json().next_page_token} response status is 200`]: (r) => r.status === 200,
            [`GET /v1alpha/source-connectors?page_size=1&page_token=${pageRes.json().next_page_token} response source_connectors size 1`]: (r) => r.json().source_connectors.length === 1,
        });

        check(http.request("GET", `${connectorHost}/v1alpha/source-connectors?page_size=1&view=VIEW_BASIC`, null, { headers: genAuthHeader(data.userAccessToken, "application/json") }), {
            "GET /v1alpha/source-connectors?page_size=1&view=VIEW_BASIC response status 200": (r) => r.status === 200,
            "GET /v1alpha/source-connectors?page_size=1&view=VIEW_BASIC response source_connectors[0]connector.configuration is null": (r) => r.json().source_connectors[0].connector.configuration === null,
        });

        check(http.request("GET", `${connectorHost}/v1alpha/source-connectors?page_size=1&view=VIEW_FULL`, null, { headers: genAuthHeader(data.userAccessToken, "application/json") }), {
            "GET /v1alpha/source-connectors?page_size=1&view=VIEW_FULL response status 200": (r) => r.status === 200,
            "GET /v1alpha/source-connectors?page_size=1&view=VIEW_FULL response source_connectors[0]connector.configuration is not null": (r) => r.json().source_connectors[0].connector.configuration !== null,
            "GET /v1alpha/source-connectors?page_size=1&view=VIEW_BASIC response source_connectors[0]connector.configuration is {}": (r) => Object.keys(r.json().source_connectors[0].connector.configuration).length === 0,
        });

        check(http.request("GET", `${connectorHost}/v1alpha/source-connectors?page_size=1`, null, { headers: genAuthHeader(data.userAccessToken, "application/json") }), {
            "GET /v1alpha/source-connectors?page_size=1 response status 200": (r) => r.status === 200,
            "GET /v1alpha/source-connectors?page_size=1 response source_connectors[0]connector.configuration is null": (r) => r.json().source_connectors[0].connector.configuration === null
        });

        check(http.request("GET", `${connectorHost}/v1alpha/source-connectors?page_size=${limitedRecords.json().total_size}`, null, { headers: genAuthHeader(data.userAccessToken, "application/json") }), {
            [`GET /v1alpha/source-connectors?page_size=${limitedRecords.json().total_size} response status 200`]: (r) => r.status === 200,
            [`GET /v1alpha/source-connectors?page_size=${limitedRecords.json().total_size} response next_page_token is empty`]: (r) => r.json().next_page_token === ""
        });

        // Delete the destination connectors
        for (const reqBody of reqBodies) {
            check(http.request("DELETE", `${connectorHost}/v1alpha/source-connectors/${reqBody.id}`, null, { headers: genAuthHeader(data.userAccessToken, "application/json") }), {
                [`DELETE /v1alpha/source-connectors x${reqBodies.length} response status is 204`]: (r) => r.status === 204,
            });
        }
    });
}

export function CheckGet(data) {

    group("Connector API: Get source connectors by ID", () => {

        var httpSrcConnector = {
            "id": "source-http",
            "source_connector_definition": constant.httpSrcDefRscName,
            "connector": {
                "configuration": {}
            }
        }

        var resHTTP = http.request("POST", `${connectorHost}/v1alpha/source-connectors`,
            JSON.stringify(httpSrcConnector), {
            headers: genAuthHeader(data.userAccessToken, "application/json"),
        })

        check(http.request("GET", `${connectorHost}/v1alpha/source-connectors/${resHTTP.json().source_connector.id}`, null, { headers: genAuthHeader(data.userAccessToken, "application/json") }), {
            [`GET /v1alpha/source-connectors/${resHTTP.json().source_connector.id} response status 200`]: (r) => r.status === 200,
            [`GET /v1alpha/source-connectors/${resHTTP.json().source_connector.id} response connector id`]: (r) => r.json().source_connector.id === httpSrcConnector.id,
            [`GET /v1alpha/source-connectors/${resHTTP.json().source_connector.id} response connector source_connector_definition`]: (r) => r.json().source_connector.source_connector_definition === constant.httpSrcDefRscName,
        });

        check(http.request("DELETE", `${connectorHost}/v1alpha/source-connectors/${resHTTP.json().source_connector.id}`, null, { headers: genAuthHeader(data.userAccessToken, "application/json") }), {
            [`DELETE /v1alpha/source-connectors/${resHTTP.json().source_connector.id} response status 204`]: (r) => r.status === 204,
        });

    });
}

export function CheckUpdate(data) {

    group("Connector API: Update source connectors", () => {

        var gRPCSrcConnector = {
            "id": "source-grpc",
            "source_connector_definition": constant.gRPCSrcDefRscName,
            "connector": {
                "configuration": {}
            }
        }

        check(http.request(
            "POST",
            `${connectorHost}/v1alpha/source-connectors`,
            JSON.stringify(gRPCSrcConnector), {
            headers: genAuthHeader(data.userAccessToken, "application/json"),
        }), {
            "POST /v1alpha/source-connectors response status for creating gRPC source connector 201": (r) => r.status === 201,
        });

        gRPCSrcConnector.connector.description = randomString(20)
        check(http.request(
            "PATCH",
            `${connectorHost}/v1alpha/source-connectors/${gRPCSrcConnector.id}`,
            JSON.stringify(gRPCSrcConnector), {
            headers: genAuthHeader(data.userAccessToken, "application/json"),
        }), {
            [`PATCH /v1alpha/source-connectors/${gRPCSrcConnector.id} response status for updating gRPC source connector 422`]: (r) => r.status === 422,
        });

        check(http.request("DELETE", `${connectorHost}/v1alpha/source-connectors/${gRPCSrcConnector.id}`, null, { headers: genAuthHeader(data.userAccessToken, "application/json") }), {
            [`DELETE /v1alpha/source-connectors/${gRPCSrcConnector.id} response status 204`]: (r) => r.status === 204,
        });

    });

}

export function CheckDelete(data) {

    group("Connector API: Delete source connectors", () => {

        check(http.request("POST", `${connectorHost}/v1alpha/source-connectors`,
            JSON.stringify({
                "id": "source-http",
                "source_connector_definition": "source-connector-definitions/source-http",
                "connector": {
                    "configuration": {}
                }
            }), { headers: { "Content-Type": "application/json" } }), {
            "POST /v1alpha/source-connectors response status for creating HTTP source connector 201": (r) => r.status === 201,
        })

        check(http.request("POST", `${connectorHost}/v1alpha/destination-connectors`,
            JSON.stringify({
                "id": "destination-http",
                "destination_connector_definition": "destination-connector-definitions/destination-http",
                "connector": {
                    "configuration": {}
                }
            }), { headers: { "Content-Type": "application/json" } }), {
            "POST /v1alpha/destination-connectors response status for creating HTTP destination connector 201": (r) => r.status === 201,
        })

        check(http.request("POST", `${modelHost}/v1alpha/models`, JSON.stringify({
            "id": "dummy-cls",
            "model_definition": "model-definitions/github",
            "configuration": {
                "repository": "instill-ai/model-dummy-cls"
            },
        }), { headers: { "Content-Type": "application/json" } }), {
            "POST /v1alpha/models cls response status": (r) => r.status === 201,
        })

        const detSyncRecipe = {
            recipe: {
                source: "source-connectors/source-http",
                model_instances: [`models/dummy-cls/instances/v1.0`],
                destination: "destination-connectors/destination-http"
            },
        };

        // Create a pipeline
        const pipelineID = randomString(5)
        check(http.request("POST", `${pipelineHost}/v1alpha/pipelines`,
            JSON.stringify(Object.assign({
                id: pipelineID,
                description: randomString(10),
            },
                detSyncRecipe
            )), {
            headers: {
                "Content-Type": "application/json",
            },
        }), {
            "POST /v1alpha/pipelines response status is 201": (r) => r.status === 201,
        })

        // Cannot delete source connector due to pipeline occupancy
        check(http.request("DELETE", `${connectorHost}/v1alpha/source-connectors/source-http`, null, { headers: genAuthHeader(data.userAccessToken, "application/json") }), {
            [`DELETE /v1alpha/source-connectors/source-http response status 422`]: (r) => r.status === 422,
            [`DELETE /v1alpha/source-connectors/source-http response error msg not nil`]: (r) => r.json() != {},
        });

        // Cannot delete destination connector due to pipeline occupancy
        check(http.request("DELETE", `${connectorHost}/v1alpha/destination-connectors/destination-http`, null, { headers: genAuthHeader(data.userAccessToken, "application/json") }), {
            [`DELETE /v1alpha/destination-connectors/destination-http response status 422`]: (r) => r.status === 422,
            [`DELETE /v1alpha/destination-connectors/source-http response error msg not nil`]: (r) => r.json() != {},
        });

        // Cannot delete model due to pipeline occupancy
        check(http.request("DELETE", `${modelHost}/v1alpha/models/dummy-cls`, null, { headers: genAuthHeader(data.userAccessToken, "application/json") }), {
            [`DELETE /v1alpha/models/dummy-cls response status is 204`]: (r) => r.status === 422,
            [`DELETE /v1alpha/models/dummy-cls response error msg not nil`]: (r) => r.json() != {},
        });

        check(http.request("DELETE", `${pipelineHost}/v1alpha/pipelines/${pipelineID}`, null, { headers: genAuthHeader(data.userAccessToken, "application/json") }), {
            [`DELETE /v1alpha/pipelines/${pipelineID} response status is 204`]: (r) => r.status === 204,
        });

        // Can delete source connector now
        check(http.request("DELETE", `${connectorHost}/v1alpha/source-connectors/source-http`, null, { headers: genAuthHeader(data.userAccessToken, "application/json") }), {
            [`DELETE /v1alpha/source-connectors/source-http response status 204`]: (r) => r.status === 204,
        });

        // Can delete destination connector now
        check(http.request("DELETE", `${connectorHost}/v1alpha/destination-connectors/destination-http`, null, { headers: genAuthHeader(data.userAccessToken, "application/json") }), {
            [`DELETE /v1alpha/destination-connectors/destination-http response status 204`]: (r) => r.status === 204,
        });

        // Can delete model now
        check(http.request("DELETE", `${modelHost}/v1alpha/models/dummy-cls`, null, { headers: genAuthHeader(data.userAccessToken, "application/json") }), {
            [`DELETE /v1alpha/models/dummy-cls response status is 204`]: (r) => r.status === 204,
        });

    });
}

export function CheckLookUp(data) {

    group("Connector API: Look up source connectors by UID", () => {

        var httpSrcConnector = {
            "id": "source-http",
            "source_connector_definition": constant.httpSrcDefRscName,
            "connector": {
                "configuration": {}
            }
        }

        var resHTTP = http.request("POST", `${connectorHost}/v1alpha/source-connectors`,
            JSON.stringify(httpSrcConnector), {
            headers: genAuthHeader(data.userAccessToken, "application/json"),
        })

        check(http.request("GET", `${connectorHost}/v1alpha/source-connectors/${resHTTP.json().source_connector.uid}/lookUp`, null, { headers: genAuthHeader(data.userAccessToken, "application/json") }), {
            [`GET /v1alpha/source-connectors/${resHTTP.json().source_connector.uid}/lookUp response status 200`]: (r) => r.status === 200,
            [`GET /v1alpha/source-connectors/${resHTTP.json().source_connector.uid}/lookUp response connector uid`]: (r) => r.json().source_connector.uid === resHTTP.json().source_connector.uid,
            [`GET /v1alpha/source-connectors/${resHTTP.json().source_connector.uid}/lookUp response connector source_connector_definition`]: (r) => r.json().source_connector.source_connector_definition === constant.httpSrcDefRscName,
        });

        check(http.request("DELETE", `${connectorHost}/v1alpha/source-connectors/${resHTTP.json().source_connector.id}`, null, { headers: genAuthHeader(data.userAccessToken, "application/json") }), {
            [`DELETE /v1alpha/source-connectors/${resHTTP.json().source_connector.id} response status 204`]: (r) => r.status === 204,
        });

    });
}

export function CheckState(data) {

    group("Connector API: Change state source connectors", () => {
        var httpSrcConnector = {
            "id": "source-http",
            "source_connector_definition": constant.httpSrcDefRscName,
            "connector": {
                "configuration": {}
            }
        }

        var resHTTP = http.request("POST", `${connectorHost}/v1alpha/source-connectors`,
            JSON.stringify(httpSrcConnector), {
            headers: genAuthHeader(data.userAccessToken, "application/json"),
        })

        check(http.request("POST", `${connectorHost}/v1alpha/source-connectors/${resHTTP.json().source_connector.id}/connect`, null, {
            headers: { "Content-Type": "application/json" }
        }), {
            [`POST /v1alpha/source-connectors/${resHTTP.json().source_connector.id}/connect response status 200`]: (r) => r.status === 200,
        });

        check(http.request("POST", `${connectorHost}/v1alpha/source-connectors/${resHTTP.json().source_connector.id}/disconnect`, null, {
            headers: { "Content-Type": "application/json" }
        }), {
            [`POST /v1alpha/source-connectors/${resHTTP.json().source_connector.id}/disconnect response status 422`]: (r) => r.status === 422,
        });

        check(http.request("DELETE", `${connectorHost}/v1alpha/source-connectors/${resHTTP.json().source_connector.id}`, null, { headers: genAuthHeader(data.userAccessToken, "application/json") }), {
            [`DELETE /v1alpha/source-connectors/${resHTTP.json().source_connector.id} response status 204`]: (r) => r.status === 204,
        });

    });

}

export function CheckRename(data) {

    group("Connector API: Rename source connectors", () => {
        var httpSrcConnector = {
            "id": "source-http",
            "source_connector_definition": constant.httpSrcDefRscName,
            "connector": {
                "configuration": {}
            }
        }

        var resHTTP = http.request("POST", `${connectorHost}/v1alpha/source-connectors`,
            JSON.stringify(httpSrcConnector), {
            headers: genAuthHeader(data.userAccessToken, "application/json"),
        })

        check(http.request("POST", `${connectorHost}/v1alpha/source-connectors/${resHTTP.json().source_connector.id}/rename`,
            JSON.stringify({
                "new_source_connector_id": "some-id-not-http"
            }), {
            headers: { "Content-Type": "application/json" }
        }), {
            [`POST /v1alpha/source-connectors/${resHTTP.json().source_connector.id}/rename response status 422`]: (r) => r.status === 422,
        });

        check(http.request("DELETE", `${connectorHost}/v1alpha/source-connectors/${resHTTP.json().source_connector.id}`, null, { headers: genAuthHeader(data.userAccessToken, "application/json") }), {
            [`DELETE /v1alpha/source-connectors/${resHTTP.json().source_connector.id} response status 204`]: (r) => r.status === 204,
        });
    });

}
