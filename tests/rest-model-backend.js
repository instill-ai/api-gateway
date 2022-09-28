import http from "k6/http";
import {
  check,
  group,
} from "k6";
import {
  randomString
} from "https://jslib.k6.io/k6-utils/1.1.0/index.js";

import * as createModel from "./model-backend/rest_create_model.js"
import * as queryModel from "./model-backend/rest_query_model.js"
import * as inferModel from "./model-backend/rest_infer_model.js"
import * as deployModel from "./model-backend/rest_deploy_model.js"
import * as publishModel from "./model-backend/rest_publish_model.js"
import * as updateModel from "./model-backend/rest_update_model.js"
import * as queryModelDefinition from "./model-backend/rest_query_model_definition.js"
import * as queryModelInstance from "./model-backend/rest_query_model_instance.js"
import * as getModelCard from "./model-backend/rest_model_card.js"

import {
  genAuthHeader,
  generateUserToken,
  deleteUser,
  deleteUserClients,
} from "./helpers.js";

const hydraHost = "https://127.0.0.1:4445";
const kratosHost = "https://127.0.0.1:4434";
const apiHost = "https://127.0.0.1:8000"; // set as api gateway url, the same as HYDRA_AUDIENCE in .env.dev

const testUserEmail = `test_${randomString(10)}@foo.bar`;

export let options = {
  insecureSkipTLSVerify: true,
  thresholds: {
    checks: ["rate == 1.0"],
  },
};

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
   * Model API - API CALLS
   */

  // Health check
  {
    group("Model API: health check", () => {
      check(http.request("GET", `${apiHost}/v1alpha/health/model`), {
        "GET /v1alpha/health/model response status is 200": (r) => r.status === 200,
      });
    });
  }

  // Create Model API
  createModel.CreateModelFromLocal(data)
  createModel.CreateModelFromGitHub(data)

  // Query Model API
  queryModel.GetModel(data)
  queryModel.ListModel(data)
  queryModel.LookupModel(data)

  // Deploy/Undeploy Model API
  deployModel.DeployUndeployModel(data)

  // Infer Model API
  inferModel.InferModel(data)

  // Publish/Unpublish Model API
  publishModel.PublishUnpublishModel(data)

  // Update Model API
  updateModel.UpdateModel(data)

  // Query Model Definition API
  queryModelDefinition.GetModelDefinition(data)
  queryModelDefinition.ListModelDefinition(data)

  // Query Model Instance API
  queryModelInstance.GetModelInstance(data)
  queryModelInstance.ListModelInstance(data)
  queryModelInstance.LookupModelInstance(data)

  // Get model card
  getModelCard.GetModelCard(data)
}

export function teardown(data) {
  deleteUserClients(hydraHost, data.user.id);
  deleteUser(kratosHost, data.user.id);
  group("Model API: Delete all models created by this test", () => {
    for (const model of http
      .request("GET", `${apiHost}/v1alpha/models`, null, {
        headers: genAuthHeader(data.userAccessToken, "application/json"),
      })
      .json("models")) {
      check(
        http.request("DELETE", `${apiHost}/v1alpha/models/${model.id}`, null, {
          headers: genAuthHeader(data.userAccessToken, "application/json"),
        }),
        {
          [`DELETE /v1alpha/models/${model.id} response status is 204`]: (r) =>
            r.status === 204,
        }
      );
    }
  });
}
