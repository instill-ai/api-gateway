import http from "k6/http";
import {sleep, check, group, fail} from "k6";
import {FormData} from "https://jslib.k6.io/formdata/0.0.2/index.js";
import {randomString} from "https://jslib.k6.io/k6-utils/1.1.0/index.js";
import {URL} from "https://jslib.k6.io/url/1.0.0/index.js";

import {
  generateUserToken,
  deleteUser,
  deleteUserClients,
  genAuthHeader,
  genNoAuthHeader,
} from "./helpers.js";

import * as pipelineConstants from "./pipeline-backend-constants.js";

const hydraHost = "https://127.0.0.1:4445";
const kratosHost = "https://127.0.0.1:4434";
const apiHost = "https://127.0.0.1:8000"; // set as api gateway url, the same as HYDRA_AUDIENCE in .env.dev

const testFooBarUserEmail = `test_${randomString(10)}@foo.bar`;
const testInstillUserEmail = `test_${randomString(10)}@instill.tech`;

export let options = {
  insecureSkipTLSVerify: true,
  thresholds: {
    checks: ["rate == 1.0"],
  },
};

export function setup() {
  // Import fooBar test user and generate a fooBar user token
  const [testFooBarUser, fooBarUserAccessToken] = generateUserToken(
    kratosHost,
    hydraHost,
    apiHost,
    testFooBarUserEmail
  );

  // Import Instill test user and generate a Instill user token
  const [testInstillUser, instillUserAccessToken] = generateUserToken(
    kratosHost,
    hydraHost,
    apiHost,
    testInstillUserEmail
  );

  return {
    fooBar: {
      user: testFooBarUser,
      userAccessToken: fooBarUserAccessToken,
    },
    instill: {
      user: testInstillUser,
      userAccessToken: instillUserAccessToken,
    },
  };
}

const gopherImg = open(`${__ENV.TEST_FOLDER_ABS_PATH}/gopher.png`, "b");

export default function (data) {
  let instillResp;
  let fooBarResp;

  const testHeaders = {
    InstillUserTokenAuthHeader: genAuthHeader(
      data.instill.userAccessToken,
      "application/json"
    ),
    fooBarUserTokenAuthHeader: genAuthHeader(
      data.fooBar.userAccessToken,
      "application/json"
    ),
    missingAuthHeader: genNoAuthHeader("application/json"),
  };

  /*
   * Pipelines API - API CALLS
   */

  // Health check
  {
    group("Pipelines API: Health check", () => {
      check(http.request("GET", `${apiHost}/health/pipeline`), {
        "GET /health/pipelines response status is 200": (r) => r.status === 200,
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
      pipelineConstants.fooBarDetectionRecipe
    );
    let createUnmarshallEntity = Object.assign(
      {
        name: randomString(100),
        description: randomString(512),
        active: true,
      },
      pipelineConstants.unmarshallRecipe
    );
    group("Pipelines API: Create pipeline with Instill model", () => {
      instillResp = http.request(
        "POST",
        `${apiHost}/pipelines`,
        JSON.stringify(createInstillPipelineEntity),
        {
          headers: testHeaders.InstillUserTokenAuthHeader,
        }
      );
      check(instillResp, {
        "POST /pipelines response status is 201": (r) => r.status === 201,
        "POST /pipelines response id check": (r) => r.json("id") != null,
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
      check(
        http.request("POST", `${apiHost}/pipelines`, JSON.stringify({}), {
          headers: testHeaders.InstillUserTokenAuthHeader,
        }),
        {
          "POST /pipelines response status is 400": (r) => r.status === 400,
        }
      );
      check(
        http.request("POST", `${apiHost}/pipelines`, null, {
          headers: testHeaders.InstillUserTokenAuthHeader,
        }),
        {
          "POST /pipelines response status is 400": (r) => r.status === 400,
        }
      );
      check(
        http.request(
          "POST",
          `${apiHost}/pipelines`,
          JSON.stringify(
            Object.assign({name: randomString(100), description: randomString(512), active: true,}, pipelineConstants.classificationRecipe)
          ),
          {
            headers: genNoAuthHeader("application/json"),
          }
        ),
        {
          "POST /pipelines response status is 401": (r) => r.status === 401,
        }
      );
    });

    // No need to create fooBar model
    // group("Pipelines API: Create pipeline with FooBar model", () => {
    //   fooBarResp = http.request(
    //     "POST",
    //     `${apiHost}/pipelines`,
    //     JSON.stringify(createFooBarPipelineEntity),
    //     {
    //       headers: testHeaders.fooBarUserTokenAuthHeader,
    //     }
    //   );
    //   check(InstillResp, {
    //     "POST /pipelines response status is 201": (r) => r.status === 201,
    //     "POST /pipelines response id check": (r) => r.json("id") != null,
    //   });
    // });

    group("Pipelines API: List pipelines", () => {
      check(
        http.request("GET", `${apiHost}/pipelines`, null, {
          headers: testHeaders.InstillUserTokenAuthHeader,
        }),
        {
          "GET /pipelines (url) response for Instill status is 200": (r) =>
            r.status === 200,
          "GET /pipelines response for Instill contents >= 1": (r) =>
            r.json("contents").length >= 1,
          "GET /pipelines response for Instill contents should not have recipe":
            (r) => r.json("contents")[0].recipe === undefined || r.json("contents")[0].recipe == null,
        }
      );
      // No need to check fooBar model
      // check(
      //   http.request("GET", `${apiHost}/pipelines`, null, {
      //     headers: testHeaders.fooBarUserTokenAuthHeader,
      //   }),
      //   {
      //     "GET /pipelines (url) response for FooBar status is 200": (r) =>
      //       r.status === 200,
      //     // This two test assetions are commented intentionally due because we don't have self and kind now
      //     "GET /pipelines response for FooBar contents >= 1": (r) =>
      //       r.json("contents").length >= 1,
      //     "GET /pipelines response for FooBar contents should not have recipe": (
      //       r
      //     ) => r.json("contents")[0].recipe === undefined || r.json("contents")[0].recipe == null,
      //   }
      // );
      const url = new URL(`${apiHost}/pipelines`);
      url.searchParams.append("view", "WITH_RECIPE");
      check(
        http.request("GET", url.toString(), null, {
          headers: testHeaders.InstillUserTokenAuthHeader,
        }),
        {
          "GET /pipelines (url) response for Instill status is 200": (r) =>
            r.status === 200,
          "GET /pipelines response for Instill contents should have recipe": (r) => 
            r.json("contents")[0].recipe != null,
        }
      );
      check(
        http.request("GET", `${apiHost}/pipelines`, null, {
          headers: testHeaders.missingAuthHeader,
        }),
        {
          "GET /pipelines (url) response status is 401": (r) =>
            r.status === 401,
        }
      );
    });

    group("Pipelines API: Get a pipeline", () => {
      check(
        http.request(
          "GET",
          `${apiHost}/pipelines/${instillResp.json("name")}`,
          null,
          {
            headers: testHeaders.InstillUserTokenAuthHeader,
          }
        ),
        {
          [`GET /pipelines/${instillResp.json("name")} response status is 200`]: (
            r
          ) => r.status === 200,
          [`GET /pipelines/${instillResp.json("name")} response name`]: (r) =>
            r.json("name") === createInstillPipelineEntity.name,
          [`GET /pipelines/${instillResp.json("name")} response description`]: (r) =>
            r.json("description") === createInstillPipelineEntity.description,
          [`GET /pipelines/${instillResp.json("name")} response id`]: (r) =>
            r.json("id") != null,
          [`GET /pipelines/${instillResp.json("name")} response recipe`]: (r) =>
            r.json("recipe") != null,
        }
      );
      check(
        http.request(
          "GET",
          `${apiHost}/pipelines/${instillResp.json("name")}`,
          null,
          {
            headers: genNoAuthHeader("application/json"),
          }
        ),
        {
          [`GET /pipelines/${instillResp.json("name")} response status is 401`]: (
            r
          ) => r.status === 401,
        }
      );
      check(
        http.request("GET", `${apiHost}/pipelines/non_exist_id`, null, {
          headers: testHeaders.InstillUserTokenAuthHeader,
        }),
        {
          "GET /pipelines/non_exist_id response status is 404": (r) =>
            r.status === 404,
        }
      );
    });

    let updateInstillPipelineEntity = Object.assign(
      {
        name: randomString(100),
        description: randomString(512),
        active: true,
      },
      pipelineConstants.cocoDetectionRecipe
    );
    let updateFooBarPipelineEntity = Object.assign(
      {
        name: randomString(100),
        description: randomString(512),
        active: true,
      },
      pipelineConstants.fooBarDetectionRecipe
    );
    group("Pipelines API: Update a pipeline", () => {
      check(
        http.request(
          "PATCH",
          `${apiHost}/pipelines/${instillResp.json("name")}`,
          JSON.stringify(updateInstillPipelineEntity),
          {
            headers: testHeaders.InstillUserTokenAuthHeader,
          }
        ),
        {
          [`PATCH /pipelines/${instillResp.json("name")} response status is 200`]:
            (r) => r.status === 200,
          [`PATCH /pipelines/${instillResp.json("name")} response description`]: (
            r
          ) =>
            r.json("description") === updateInstillPipelineEntity.description,
          [`PATCH /pipelines/${instillResp.json("name")} response id`]: (r) =>
            r.json("id") != null,
          [`PATCH /pipelines/${instillResp.json("name")} response recipe`]: (r) =>
            r.json("recipe") != null,
          [`PATCH /pipelines/${instillResp.json("name")} response duration`]: (r) => 
            r.json("duration") != null,
        }
      );
      // This test case is commented intentionally to wait for model-backend's refactoring
      // check(
      //   http.request(
      //     "PATCH",
      //     `${apiHost}/pipelines/${InstillResp.json("name")}`,
      //     JSON.stringify(updateFooBarPipelineEntity),
      //     {
      //       headers: testHeaders.InstillUserTokenAuthHeader,
      //     }
      //   ),
      //   {
      //     [`PATCH /pipelines/${InstillResp.json("name")} response status is 422`]:
      //       (r) => r.status === 422,
      //   }
      // );
      check(
        http.request(
          "PATCH",
          `${apiHost}/pipelines/${instillResp.json("name")}`,
          null,
          {
            headers: testHeaders.InstillUserTokenAuthHeader,
          }
        ),
        {
          [`PATCH /pipelines/${instillResp.json("name")} response status is 200`]:
            (r) => r.status === 200,
        }
      );
      check(
        http.request(
          "PATCH",
          `${apiHost}/pipelines/${instillResp.json("name")}`,
          JSON.stringify(updateInstillPipelineEntity),
          {
            headers: testHeaders.missingAuthHeader,
          }
        ),
        {
          [`PATCH /pipelines/${instillResp.json("name")} response status is 401`]:
            (r) => r.status === 401,
        }
      );
      check(
        http.request(
          "PATCH",
          `${apiHost}/pipelines/non_exist_id`,
          JSON.stringify(updateInstillPipelineEntity),
          {
            headers: testHeaders.InstillUserTokenAuthHeader,
          }
        ),
        {
          "PATCH /pipelines/non_exist_id response status is 404": (r) =>
            r.status === 404,
        }
      );
    });

    group("Pipelines API: Trigger a pipeline with classification model", () => {
      check(
        http.request(
          "PATCH",
          `${apiHost}/pipelines/${instillResp.json("name")}`,
          JSON.stringify(pipelineConstants.classificationRecipe),
          {
            headers: testHeaders.InstillUserTokenAuthHeader,
          }
        ),
        {
          [`PATCH /pipelines/${instillResp.json("name")} response status is 200`]:
            (r) => r.status === 200,
        }
      );
      // url data
      check(
        http.request(
          "POST",
          `${apiHost}/pipelines/${instillResp.json("name")}/outputs`,
          JSON.stringify(pipelineConstants.triggerPipelineJSONUrl),
          {
            headers: testHeaders.InstillUserTokenAuthHeader,
          }
        ),
        {
          [`POST /pipelines/${instillResp.json(
            "id"
          )}/outputs (url) response status is 200`]: (r) => r.status === 200,
          [`POST /pipelines/${instillResp.json(
            "id"
          )}/outputs (url) response kind`]: (r) =>
            r.json("kind") === "Collection",
          [`POST /pipelines/${instillResp.json(
            "id"
          )}/outputs (url) response duration`]: (r) =>
            r.json("duration") != null,
          [`POST /pipelines/${instillResp.json(
            "id"
          )}/outputs (url) response contents.length`]: (r) =>
            r.json("contents").length === 1,
          [`POST /pipelines/${instillResp.json(
            "id"
          )}/outputs (url) response contents[0].kind`]: (r) =>
            r.json("contents")[0].kind === "ClassificationOutput",
          [`POST /pipelines/${instillResp.json(
            "id"
          )}/outputs (url) response contents[0].contents.length`]: (r) =>
            r.json("contents")[0].contents.length === 5,
          [`POST /pipelines/${instillResp.json(
            "id"
          )}/outputs (url) response contents[0].contents[0].kind`]: (r) =>
            r.json("contents")[0].contents[0].kind === "CategoryPrediction",
          [`POST /pipelines/${instillResp.json(
            "id"
          )}/outputs (url) response contents[0].contents[0].category`]: (r) =>
            r.json("contents")[0].contents[0].category === "golden retriever",
          [`POST /pipelines/${instillResp.json(
            "id"
          )}/outputs (url) response contents[0].contents[0].score`]: (r) =>
            r.json("contents")[0].contents[0].score != null,
        }
      );

      // base64 data
      check(
        http.request(
          "POST",
          `${apiHost}/pipelines/${instillResp.json("name")}/outputs`,
          JSON.stringify(pipelineConstants.triggerPipelineJSONBase64),
          {
            headers: testHeaders.InstillUserTokenAuthHeader,
          }
        ),
        {
          [`POST /pipelines/${instillResp.json(
            "id"
          )}/outputs (base64) response status is 200`]: (r) => r.status === 200,
        }
      );

      // multipart data
      const fd = new FormData();
      fd.append("contents", http.file(gopherImg));
      check(
        http.request(
          "POST",
          `${apiHost}/pipelines/${instillResp.json("name")}/upload/outputs`,
          fd.body(),
          {
            headers: genAuthHeader(
              data.instill.userAccessToken,
              `multipart/form-data; boundary=${fd.boundary}`
            ),
          }
        ),
        {
          [`POST /pipelines/${instillResp.json(
            "id"
          )}/outputs (multipart) response status is 200`]: (r) =>
            r.status === 200,
        }
      );
    });

    group("Pipelines API: Trigger a pipeline with coco detection model", () => {
      check(
        http.request(
          "PATCH",
          `${apiHost}/pipelines/${instillResp.json("name")}`,
          JSON.stringify(pipelineConstants.cocoDetectionRecipe),
          {
            headers: testHeaders.InstillUserTokenAuthHeader,
          }
        ),
        {
          [`PATCH /pipelines/${instillResp.json("name")} response status is 200`]:
            (r) => r.status === 200,
        }
      );
      // url data
      check(
        http.request(
          "POST",
          `${apiHost}/pipelines/${instillResp.json("name")}/outputs`,
          JSON.stringify(pipelineConstants.triggerPipelineJSONUrl),
          {
            headers: testHeaders.InstillUserTokenAuthHeader,
          }
        ),
        {
          [`POST /pipelines/${instillResp.json(
            "id"
          )}/outputs (url) response status is 200`]: (r) => r.status === 200,
          [`POST /pipelines/${instillResp.json(
            "id"
          )}/outputs (url) response kind`]: (r) =>
            r.json("kind") === "Collection",
          [`POST /pipelines/${instillResp.json(
            "id"
          )}/outputs (url) response duration`]: (r) =>
            r.json("duration") != null,
          [`POST /pipelines/${instillResp.json(
            "id"
          )}/outputs (url) response contents.length`]: (r) =>
            r.json("contents").length === 1,
          [`POST /pipelines/${instillResp.json(
            "id"
          )}/outputs (url) response contents[0].kind`]: (r) =>
            r.json("contents")[0].kind === "DetectionOutput",
          [`POST /pipelines/${instillResp.json(
            "id"
          )}/outputs (url) response contents[0].contents[0].kind`]: (r) =>
            r.json("contents")[0].contents[0].kind === "BoundingBoxPrediction",
          [`POST /pipelines/${instillResp.json(
            "id"
          )}/outputs (url) response contents[0].contents[0].category`]: (r) =>
            r.json("contents")[0].contents[0].category === "dog",
          [`POST /pipelines/${instillResp.json(
            "id"
          )}/outputs (url) response contents[0].contents[0].score`]: (r) =>
            r.json("contents")[0].contents[0].score != null,
          [`POST /pipelines/${instillResp.json(
            "id"
          )}/outputs (url) response contents[0].contents[0].box.top`]: (r) =>
            r.json("contents")[0].contents[0].box.top != null,
          [`POST /pipelines/${instillResp.json(
            "id"
          )}/outputs (url) response contents[0].contents[0].box.left`]: (r) =>
            r.json("contents")[0].contents[0].box.left != null,
          [`POST /pipelines/${instillResp.json(
            "id"
          )}/outputs (url) response contents[0].contents[0].box.width`]: (r) =>
            r.json("contents")[0].contents[0].box.width != null,
          [`POST /pipelines/${instillResp.json(
            "id"
          )}/outputs (url) response contents[0].contents[0].box.height`]: (r) =>
            r.json("contents")[0].contents[0].box.height != null,
        }
      );

      // base64 data
      check(
        http.request(
          "POST",
          `${apiHost}/pipelines/${instillResp.json("name")}/outputs`,
          JSON.stringify(pipelineConstants.triggerPipelineJSONBase64),
          {
            headers: testHeaders.InstillUserTokenAuthHeader,
          }
        ),
        {
          [`POST /pipelines/${instillResp.json(
            "id"
          )}/outputs (base64) response status is 200`]: (r) => r.status === 200,
        }
      );

      // multipart data
      const fd = new FormData();
      fd.append("contents", http.file(gopherImg));
      check(
        http.request(
          "POST",
          `${apiHost}/pipelines/${instillResp.json("name")}/upload/outputs`,
          fd.body(),
          {
            headers: genAuthHeader(
              data.instill.userAccessToken,
              `multipart/form-data; boundary=${fd.boundary}`
            ),
          }
        ),
        {
          [`POST /pipelines/${instillResp.json(
            "id"
          )}/outputs (multipart) response status is 200`]: (r) =>
            r.status === 200,
        }
      );
    });
  }

  sleep(1);
}

export function teardown(data) {
  group("Pipeline API: Delete all pipelines created by this test", () => {
    // delete all Instill pipelines
    for (const pipeline of http
      .request("GET", `${apiHost}/pipelines`, null, {
        headers: genAuthHeader(
          data.instill.userAccessToken,
          "application/json"
        ),
      })
      .json("contents")) {
      check(pipeline, {
        "GET /clients response contents[*] id": (c) => c.id != null,
      });

      check(
        http.request("DELETE", `${apiHost}/pipelines/${pipeline.name}`, null, {
          headers: genAuthHeader(
            data.instill.userAccessToken,
            "application/json"
          ),
        }),
        {
          [`DELETE /pipelines/${pipeline.name} response status is 204`]: (r) =>
            r.status === 204,
        }
      );
    }
    // delete all fooBar pipelines
    for (const pipeline of http
      .request("GET", `${apiHost}/pipelines`, null, {
        headers: genAuthHeader(data.fooBar.userAccessToken, "application/json"),
      })
      .json("contents")) {
      check(pipeline, {
        "GET /clients response contents[*] id": (c) => c.id != null,
      });

      check(
        http.request("DELETE", `${apiHost}/pipelines/${pipeline.name}`, null, {
          headers: genAuthHeader(data.fooBar.userAccessToken, "application/json"),
        }),
        {
          [`DELETE /pipelines/${pipeline.name} response status is 204`]: (r) =>
            r.status === 204,
        }
      );
    }
  });

  deleteUserClients(hydraHost, data.fooBar.user.id);
  deleteUser(kratosHost, data.fooBar.user.id);
  deleteUserClients(hydraHost, data.instill.user.id);
  deleteUser(kratosHost, data.instill.user.id);
}
