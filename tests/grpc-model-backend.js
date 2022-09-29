import grpc from 'k6/net/grpc';
import { check, group } from 'k6';
import {
    randomString
} from "https://jslib.k6.io/k6-utils/1.1.0/index.js";


import * as createModel from "./model-backend/grpc_create_model.js"
import * as updateModel from "./model-backend/grpc_update_model.js"
import * as queryModel from "./model-backend/grpc_query_model.js"
import * as deployModel from "./model-backend/grpc_deploy_model.js"
import * as inferModel from "./model-backend/grpc_infer_model.js"
import * as publishModel from "./model-backend/grpc_publish_model.js"
import * as queryModelInstance from "./model-backend/grpc_query_model_instance.js"
import * as queryModelDefinition from "./model-backend/grpc_query_model_definition.js"


const client = new grpc.Client();
client.load(['proto'], 'model_definition.proto');
client.load(['proto'], 'model.proto');
client.load(['proto'], 'model_service.proto');
client.load(['proto'], 'healthcheck.proto');

import {
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
    tlsAuth: [
        {
            domains: ['localhost'],
            cert: open('../secrets/certs/model-backend/model-backend.crt'),
            key: open('../secrets/certs/model-backend/model-backend.key'),
        },
    ],
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

export default function(data) {
    // Liveness check
    {
        group("Model API: Liveness", () => {
            client.connect('localhost:8000', {});
            const response = client.invoke('vdp.model.v1alpha.ModelService/Liveness', {});
            check(response, {
                'Status is OK': (r) => r && r.status === grpc.StatusOK,
                'Response status is SERVING_STATUS_SERVING': (r) => r && r.message.healthCheckResponse.status === "SERVING_STATUS_SERVING",
            });
        });
    }

    // Readiness check
    group("Model API: Readiness", () => {
        client.connect('localhost:8000', {});
        const response = client.invoke('vdp.model.v1alpha.ModelService/Readiness', {});
        check(response, {
            'Status is OK': (r) => r && r.status === grpc.StatusOK,
            'Response status is SERVING_STATUS_SERVING': (r) => r && r.message.healthCheckResponse.status === "SERVING_STATUS_SERVING",
        });
        client.close();
    });

    // Create model API
    createModel.CreateModel(data)

    // Update model API
    updateModel.UpdateModel(data)

    // Deploy Model API
    deployModel.DeployUndeployModel(data)

    // Query Model API
    queryModel.GetModel(data)
    queryModel.ListModel(data)
    queryModel.LookupModel(data)

    // Publish Model API
    publishModel.PublishUnPublishModel(data)

    // // Infer Model API
    inferModel.InferModel(data)

    // Query Model Instance API
    queryModelInstance.GetModelInstance(data)
    queryModelInstance.ListModelInstance(data)
    queryModelInstance.LookupModelInstance(data)

    // Query Model Definition API
    queryModelDefinition.GetModelDefinition(data)
    queryModelDefinition.ListModelDefinition(data)
};

export function teardown(data) {
    deleteUserClients(hydraHost, data.user.id);
    deleteUser(kratosHost, data.user.id);
    client.connect('localhost:8000', {});
    group("Model API: Delete all models created by this test", () => {
        for (const model of client.invoke('vdp.model.v1alpha.ModelService/ListModel', {}, {}).message.models) {
            check(client.invoke('vdp.model.v1alpha.ModelService/DeleteModel', { name: model.name }), {
                'DeleteModel model status is OK': (r) => r && r.status === grpc.StatusOK,
            });
        }
    });
    client.close();
}
