export const pipelineHost = "https://127.0.0.1:8000";
export const connectorHost = "https://127.0.0.1:8000";
export const modelHost = "https://127.0.0.1:8000";

export const dogImg = open(`${__ENV.TEST_FOLDER_ABS_PATH}/data/dog.jpg`, "b");
export const catImg = open(`${__ENV.TEST_FOLDER_ABS_PATH}/data/cat.jpg`, "b");
export const bearImg = open(`${__ENV.TEST_FOLDER_ABS_PATH}/data/bear.jpg`, "b");
export const dogRGBAImg = open(`${__ENV.TEST_FOLDER_ABS_PATH}/data/dog-rgba.png`, "b");

export const det_model = open(`${__ENV.TEST_FOLDER_ABS_PATH}/data/dummy-det-model.zip`, "b");
export const model_def_name = "model-definitions/local"
export const model_id = "dummy-det"
export const model_instance_id = "latest"

export const detSyncHTTPSingleModelInstRecipe = {
  recipe: {
    source: "source-connectors/source-http",
    model_instances: [
      `models/${model_id}/instances/${model_instance_id}`,
    ],
    destination: "destination-connectors/destination-http"
  },
};

export const detSyncGRPCSingleModelInstRecipe = {
  recipe: {
    source: "source-connectors/source-grpc",
    model_instances: [
      `models/${model_id}/instances/${model_instance_id}`,
    ],
    destination: "destination-connectors/destination-grpc"
  },
};

export const detSyncHTTPMultiModelInstRecipe = {
  recipe: {
    source: "source-connectors/source-http",
    model_instances: [
      `models/${model_id}/instances/${model_instance_id}`,
      `models/${model_id}/instances/${model_instance_id}`,
    ],
    destination: "destination-connectors/destination-http"
  },
};

export const dstCSVConnID = "some-cool-name-for-dst-csv-connector"

export const detAsyncSingleModelInstRecipe = {
  recipe: {
    source: "source-connectors/source-http",
    model_instances: [
      `models/${model_id}/instances/${model_instance_id}`
    ],
    destination:`destination-connectors/${dstCSVConnID}`
  },
};

export const detAsyncMultiModelInstRecipe = {
  recipe: {
    source: "source-connectors/source-http",
    model_instances: [
      `models/${model_id}/instances/${model_instance_id}`,
      `models/${model_id}/instances/${model_instance_id}`,
    ],
    destination:`destination-connectors/${dstCSVConnID}`
  },
};
