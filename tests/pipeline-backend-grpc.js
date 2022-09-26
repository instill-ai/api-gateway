import grpc from 'k6/net/grpc';
import { sleep, check, group, fail } from "k6";
import encoding from "k6/encoding";
import http from "k6/http";
import { FormData } from "https://jslib.k6.io/formdata/0.0.2/index.js";
import { randomString } from "https://jslib.k6.io/k6-utils/1.1.0/index.js";
import { URL } from "https://jslib.k6.io/url/1.0.0/index.js";

import {
  generateUserToken,
  deleteUser,
  deleteUserClients,
} from "./helpers.js";

import * as pipelineConstants from "./pipeline-backend-constants.js";

const hydraHost = "https://127.0.0.1:4445";
const kratosHost = "https://127.0.0.1:4434";
const apiHost = "https://127.0.0.1:8000"; // set as api gateway url, the same as HYDRA_AUDIENCE in .env.dev

const testFooBarUserEmail = `test_${randomString(10)}@foo.bar`;
const testInstillUserEmail = `test_${randomString(10)}@instill.tech`;

const client = new grpc.Client();
client.load(['definitions'], './pipeline.proto.grpcurl');

export let options = {
  insecureSkipTLSVerify: true,
  thresholds: {
    checks: ["rate == 1.0"],
  },
};

export function setup() {

  // Import Instill test user and generate a Instill user token
  const [testInstillUser, instillUserAccessToken] = generateUserToken(
    kratosHost,
    hydraHost,
    apiHost,
    testInstillUserEmail
  );

  return {
    Instill: {
      user: testInstillUser,
      userAccessToken: instillUserAccessToken,
    },
  };
}

export default (data) => {
  let instillResp;

  const InstillUserTokenAuthHeader = {
    authorization: `bearer ${data.Instill.userAccessToken}`,
  };
  const missingAuthHeader = {};

  client.connect('localhost:8000', {
    plaintext: false
  });

  /*
   * Pipelines gRPC - API CALLS
   */

  // Health check
  {
    group("Pipelines gRPC: Liveness", () => {
      check(client.invoke('instill.pipeline.Pipeline/Liveness', {}), {
        "call RPC instill.pipeline.Pipeline/Liveness response status is OK": (r) => r.status === grpc.StatusOK,
      });
    });
    group("Pipelines gRPC: Readiness", () => {
      check(client.invoke('instill.pipeline.Pipeline/Readiness', {}), {
        "call RPC instill.pipeline.Pipeline/Readiness response status is OK": (r) => r.status === grpc.StatusOK,
      });
    });
  }

  // Pipelines
  {
    let createInstillPipelineEntity = Object.assign(
      {
        name: randomString(100),
        description: randomString(512),
        active: true,
      },
      pipelineConstants.classificationRecipe
    );
    let createFooBarPipelineEntity = Object.assign(
      {
        name: randomString(100),
        description: randomString(512),
        active: true,
      },
      pipelineConstants.cuboDetectionRecipe
    );
    let createUnmarshallEntity = Object.assign(
      {
        name: randomString(100),
        description: randomString(512),
        active: true,
      },
      pipelineConstants.unmarshallRecipe
    );

    group("Pipelines gRPC: CreatePipeline with Instill model", () => {
      instillResp = client.invoke('instill.pipeline.Pipeline/CreatePipeline', createInstillPipelineEntity, { headers: InstillUserTokenAuthHeader });
      check(instillResp, {
        "call RPC instill.pipeline.Pipeline/CreatePipeline response status is OK": (r) => r.status === grpc.StatusOK,
        "call RPC instill.pipeline.Pipeline/CreatePipeline response id check": (r) => r.message.id !== undefined,
      });
      // This test case is commented intentionally to wait for model-backend's refactoring
      // check(
      //   http.request(
      //     "POST",
      //     `${apiHost}/pipelines`,
      //     JSON.stringify(createFooBarPipelineEntity),
      //     {
      //       headers: testHeaders.InstillUserTokenAuthHeader,
      //     }
      //   ),
      //   {
      //     "POST /pipelines response status is 422": (r) => r.status === 422,
      //   }
      // );
      // check(
      //   client.invoke('instill.pipeline.Pipeline/CreatePipeline', {}, { headers: InstillUserTokenAuthHeader }),
      //   {
      //     "call RPC instill.pipeline.Pipeline/CreatePipeline response status is FailedPrecondition": (r) => r.status === grpc.StatusFailedPrecondition,
      //   }
      // );
      check(
        client.invoke('instill.pipeline.Pipeline/CreatePipeline', {}, { headers: missingAuthHeader }),
        {
          "call RPC instill.pipeline.Pipeline/CreatePipeline response status is Unauthenticated": (r) => r.status === grpc.StatusUnauthenticated,
        }
      );
    });

    // No need to create cubo model
    // group("Pipelines gRPC: Create pipeline with FooBar model", () => {
    // });

    group("Pipelines gRPC: List pipelines", () => {
      check(
        client.invoke('instill.pipeline.Pipeline/ListPipelines', {}, { headers: InstillUserTokenAuthHeader }),
        {
          "call RPC instill.pipeline.Pipeline/ListPipelines response for Instill status is OK": (r) => r.status === grpc.StatusOK,
          "call RPC instill.pipeline.Pipeline/ListPipelines response for Instill contents >= 1": (r) => r.message.contents.length >= 1,
          "call RPC instill.pipeline.Pipeline/ListPipelines response for Instill contents should not have recipe":
            (r) => r.message.contents[0].recipe === undefined || r.message.contents[0].recipe == null,
        }
      );
      check(
        client.invoke('instill.pipeline.Pipeline/ListPipelines', { view: "WITH_RECIPE" }, { headers: InstillUserTokenAuthHeader }),
        {
          "call RPC instill.pipeline.Pipeline/ListPipelines response for Instill status is OK": (r) => r.status === grpc.StatusOK,
          "call RPC instill.pipeline.Pipeline/ListPipelines response for Instill contents should have recipe": (r) => r.message.contents[0].recipe !== undefined,
        }
      );
      check(
        client.invoke('instill.pipeline.Pipeline/ListPipelines', {}, { headers: missingAuthHeader }),
        {
          "call RPC instill.pipeline.Pipeline/ListPipelines response status is Unauthenticated": (r) => r.status === grpc.StatusUnauthenticated,
        }
      );
    });

    group("Pipelines gRPC: Get a pipeline", () => {
      check(
        client.invoke('instill.pipeline.Pipeline/GetPipeline', { name: instillResp.message.name }, { headers: InstillUserTokenAuthHeader }),
        {
          [`call RPC instill.pipeline.Pipeline/GetPipeline with name ${instillResp.message.name} response status is OK`]:
            (r) => r.status === grpc.StatusOK,
          [`call RPC instill.pipeline.Pipeline/GetPipeline with name ${instillResp.message.name} response name`]:
            (r) => r.message.name === createInstillPipelineEntity.name,
          [`call RPC instill.pipeline.Pipeline/GetPipeline with name ${instillResp.message.name} response description`]:
            (r) => r.message.description === createInstillPipelineEntity.description,
          [`call RPC instill.pipeline.Pipeline/GetPipeline with name ${instillResp.message.name} response id`]:
            (r) => r.message.id !== undefined,
          [`call RPC instill.pipeline.Pipeline/GetPipeline with name ${instillResp.message.name} response recipe`]:
            (r) => r.message.recipe !== undefined,
        }
      );
      check(
        client.invoke('instill.pipeline.Pipeline/GetPipeline', { name: instillResp.message.name }, { headers: missingAuthHeader }),
        {
          [`call RPC instill.pipeline.Pipeline/GetPipeline with name ${instillResp.message.name} response status is Unauthenticated`]:
            (r) => r.status === grpc.StatusUnauthenticated,
        }
      );
      // check(
      //   client.invoke('instill.pipeline.Pipeline/GetPipeline', { name: "non_exist_id" }, { headers: InstillUserTokenAuthHeader }),
      //   {
      //     "call RPC instill.pipeline.Pipeline/GetPipeline response is": (r) => { console.log(JSON.stringify(r)); return true; },
      //     "call RPC instill.pipeline.Pipeline/GetPipeline with name non_exist_id response status is NotFound": 
      //       (r) => r.status === grpc.StatusNotFound,
      //   }
      // );
    });

    let updateInstillPipelineEntity = {
      pipeline: Object.assign(
        {
          name: instillResp.message.name,
          description: randomString(512),
          active: true,
        },
        pipelineConstants.cocoDetectionRecipe
      ),
      update_mask: "description",
    };

    group("Pipelines gRPC: Update a pipeline", () => {
      check(
        client.invoke('instill.pipeline.Pipeline/UpdatePipeline',
          updateInstillPipelineEntity,
          { headers: InstillUserTokenAuthHeader }),
        {
          [`call RPC instill.pipeline.Pipeline/UpdatePipeline with name ${instillResp.message.name} response status is OK`]: (r) => r.status === grpc.StatusOK,
          [`call RPC instill.pipeline.Pipeline/UpdatePipeline with name ${instillResp.message.name} response description`]: (r) => r.message.description === updateInstillPipelineEntity.pipeline.description,
          [`call RPC instill.pipeline.Pipeline/UpdatePipeline with name ${instillResp.message.name} response id`]: (r) => r.message.id !== undefined,
          [`call RPC instill.pipeline.Pipeline/UpdatePipeline with name ${instillResp.message.name} response recipe`]: (r) => r.message.recipe !== undefined,
        }
      );
      // This test case is commented intentionally to wait for model-backend's refactoring
      // check(
      //   http.request(
      //     "PATCH",
      //     `${apiHost}/pipelines/${InstillResp.message.name}`,
      //     JSON.stringify(updateFooBarPipelineEntity),
      //     {
      //       headers: testHeaders.InstillUserTokenAuthHeader,
      //     }
      //   ),
      //   {
      //     [`PATCH /pipelines/${InstillResp.message.name} response status is 422`]:
      //       (r) => r.status === 422,
      //   }
      // );
      check(
        client.invoke('instill.pipeline.Pipeline/UpdatePipeline', { pipeline: { name: instillResp.message.name } }, { headers: InstillUserTokenAuthHeader }),
        {
          [`call RPC instill.pipeline.Pipeline/UpdatePipeline with name ${instillResp.message.name} response status is OK`]:
            (r) => r.status === grpc.StatusOK,
        }
      );
      check(
        client.invoke('instill.pipeline.Pipeline/UpdatePipeline', updateInstillPipelineEntity, { headers: missingAuthHeader }),
        {
          [`call RPC instill.pipeline.Pipeline/UpdatePipeline with name ${instillResp.message.name} response status is StatusUnauthenticated`]:
            (r) => r.status === grpc.StatusUnauthenticated,
        }
      );
      // updateInstillPipelineEntity.pipeline.id = "non_exist_id"
      // check(
      //   client.invoke('instill.pipeline.Pipeline/UpdatePipeline', updateInstillPipelineEntity, { headers: missingAuthHeader }),
      //   {
      //     "call RPC instill.pipeline.Pipeline/UpdatePipeline with name non_exist_id response status is NotFound":
      //       (r) => r.status === grpc.StatusNotFound,
      //   }
      // );
    });

    group("Pipelines gRPC: Trigger a pipeline with classification model", () => {
      check(
        client.invoke('instill.pipeline.Pipeline/UpdatePipeline',
          {
            pipeline: Object.assign({ name: instillResp.message.name }, pipelineConstants.classificationRecipe),
            update_mask: "recipe",
          },
          { headers: InstillUserTokenAuthHeader }),
        {
          [`call RPC instill.pipeline.Pipeline/UpdatePipeline with name ${instillResp.message.name} response status is OK`]:
            (r) => r.status === grpc.StatusOK,
        }
      );
      // url data
      check(
        client.invoke('instill.pipeline.Pipeline/TriggerPipeline',
          Object.assign({ name: instillResp.message.name }, pipelineConstants.triggerPipelineJSONUrl),
          { headers: InstillUserTokenAuthHeader }),
        {
          [`call RPC instill.pipeline.Pipeline/TriggerPipeline with name ${instillResp.message.name}/outputs (url) response status is OK`]: (r) => r.status === grpc.StatusOK,
          [`call RPC instill.pipeline.Pipeline/TriggerPipeline with name ${instillResp.message.name}/outputs (url) response contents.length`]: (r) => r.message.contents.length === 1,
          [`call RPC instill.pipeline.Pipeline/TriggerPipeline with name ${instillResp.message.name}/outputs (url) response contents[0].contents.length`]: (r) => r.message.contents[0].contents.length === 5,
          [`call RPC instill.pipeline.Pipeline/TriggerPipeline with name ${instillResp.message.name}/outputs (url) response contents[0].contents[0].category`]: (r) => r.message.contents[0].contents[0].category === "golden retriever",
          [`call RPC instill.pipeline.Pipeline/TriggerPipeline with name ${instillResp.message.name}/outputs (url) response contents[0].contents[0].score`]: (r) => r.message.contents[0].contents[0].score !== undefined,
        }
      );

      // base64 data
      check(
        client.invoke('instill.pipeline.Pipeline/TriggerPipeline',
          Object.assign({ name: instillResp.message.name }, pipelineConstants.triggerPipelineJSONBase64),
          { headers: InstillUserTokenAuthHeader }),
        {
          [`call RPC instill.pipeline.Pipeline/TriggerPipeline with name ${instillResp.message.name}/outputs (base64) response status is OK`]: (r) => r.status === grpc.StatusOK,
        }
      );
    });

    group("Pipelines gRPC: Trigger a pipeline with coco detection model", () => {
      check(
        client.invoke('instill.pipeline.Pipeline/UpdatePipeline',
          {
            pipeline: Object.assign({ name: instillResp.message.name }, pipelineConstants.cocoDetectionRecipe),
            update_mask: "recipe",
          },
          { headers: InstillUserTokenAuthHeader }),
        {
          [`call RPC instill.pipeline.Pipeline/UpdatePipeline with name ${instillResp.message.name} response status is OK`]:
            (r) => r.status === grpc.StatusOK,
        }
      );
      // url data
      check(
        client.invoke('instill.pipeline.Pipeline/TriggerPipeline',
          Object.assign({ name: instillResp.message.name }, pipelineConstants.triggerPipelineJSONUrl),
          { headers: InstillUserTokenAuthHeader }),
        {
          [`call RPC instill.pipeline.Pipeline/TriggerPipeline with name ${instillResp.message.name}/outputs (url) response status is OK`]: (r) => r.status === grpc.StatusOK,
          [`call RPC instill.pipeline.Pipeline/TriggerPipeline with name ${instillResp.message.name}/outputs (url) response contents.length`]: (r) => r.message.contents.length === 1,
          [`call RPC instill.pipeline.Pipeline/TriggerPipeline with name ${instillResp.message.name}/outputs (url) response contents[0].contents[0].category`]: (r) => r.message.contents[0].contents[0].category === "dog",
          [`call RPC instill.pipeline.Pipeline/TriggerPipeline with name ${instillResp.message.name}/outputs (url) response contents[0].contents[0].score`]: (r) => r.message.contents[0].contents[0].score !== undefined,
          [`call RPC instill.pipeline.Pipeline/TriggerPipeline with name ${instillResp.message.name}/outputs (url) response contents[0].contents[0].box.top`]: (r) => r.message.contents[0].contents[0].box.top !== undefined,
          [`call RPC instill.pipeline.Pipeline/TriggerPipeline with name ${instillResp.message.name}/outputs (url) response contents[0].contents[0].box.left`]: (r) => r.message.contents[0].contents[0].box.left !== undefined,
          [`call RPC instill.pipeline.Pipeline/TriggerPipeline with name ${instillResp.message.name}/outputs (url) response contents[0].contents[0].box.width`]: (r) => r.message.contents[0].contents[0].box.width !== undefined,
          [`call RPC instill.pipeline.Pipeline/TriggerPipeline with name ${instillResp.message.name}/outputs (url) response contents[0].contents[0].box.height`]: (r) => r.message.contents[0].contents[0].box.height !== undefined,
        }
      );

      // base64 data
      check(
        client.invoke('instill.pipeline.Pipeline/TriggerPipeline',
          Object.assign({ name: instillResp.message.name }, pipelineConstants.triggerPipelineJSONBase64),
          { headers: InstillUserTokenAuthHeader }),
        {
          [`call RPC instill.pipeline.Pipeline/TriggerPipeline with name ${instillResp.message.name}/outputs (base64) response status is OK`]: (r) => r.status === grpc.StatusOK,
        }
      );
    });
  }

  client.close();
  sleep(1);
}

export function teardown(data) {

  const InstillUserTokenAuthHeader = {
    authorization: `bearer ${data.Instill.userAccessToken}`,
  };

  client.connect('localhost:8000', {
    plaintext: false
  });

  group("Pipeline API: Delete all pipelines created by this test", () => {
    // delete all pipelines
    let resp = client.invoke('instill.pipeline.Pipeline/ListPipelines', {}, { headers: InstillUserTokenAuthHeader });
    check(resp, {
      "call RPC instill.pipeline.Pipeline/ListPipelines response status is OK": (r) => r.status === grpc.StatusOK,
    })
    if (resp.message.contents != null || resp.message.contents !== undefined) {
      for (const pipeline of resp.message.contents) {
        check(
          client.invoke('instill.pipeline.Pipeline/DeletePipeline', { name: pipeline.name }, { headers: InstillUserTokenAuthHeader }),
          {
            [`call RPC instill.pipeline.Pipeline/DeletePipeline with ${pipeline.name} response status is OK`]:
              (r) => r.status === grpc.StatusOK,
          }
        );
      }
    }
  });

  client.close();

  deleteUserClients(hydraHost, data.Instill.user.id);
  deleteUser(kratosHost, data.Instill.user.id);
}