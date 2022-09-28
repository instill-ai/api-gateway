export const pipelineHost = "https://127.0.0.1:8000";
export const connectorHost = "https://127.0.0.1:8000";
export const modelHost = "https://127.0.0.1:8000";

export const csvDstDefRscName = "destination-connector-definitions/destination-csv"
export const csvDstDefRscPermalink = "destination-connector-definitions/8be1cf83-fde1-477f-a4ad-318d23c9f3c6"

export const httpSrcDefRscName = "source-connector-definitions/source-http"
export const httpSrcDefRscPermalink = "source-connector-definitions/f20a3c02-c70e-4e76-8566-7c13ca11d18d"

export const gRPCSrcDefRscName = "source-connector-definitions/source-grpc"
export const gRPCSrcDefRscPermalink = "source-connector-definitions/82ca7d29-a35c-4222-b900-8d6878195e7a"

export const httpDstDefRscName = "destination-connector-definitions/destination-http"
export const httpDstDefRscPermalink = "destination-connector-definitions/909c3278-f7d1-461c-9352-87741bef11d3"

export const gRPCDstDefRscName = "destination-connector-definitions/destination-grpc"
export const gRPCDstDefRscPermalink = "destination-connector-definitions/c0e4a82c-9620-4a72-abd1-18586f2acccd"

export const mySQLDstDefRscName = "destination-connector-definitions/destination-mysql"
export const mySQLDstDefRscPermalink = "destination-connector-definitions/ca81ee7c-3163-4246-af40-094cc31e5e42"

export const csvDstConfig = {
  "destination_path": "/local/test"
};

export const clsModelInstOutputs = [
  {
    "task": "TASK_CLASSIFICATION",
    "model_instance": "models/dummy-model/instances/v1.0",
    "task_outputs": [
      {
        "index": "01GB5T5ZK9W9C2VXMWWRYM8WPS",
        "classification": {
          "category": "person",
          "score": 0.99
        }
      }
    ]
  }
]

export const detectionModelInstOutputs = [
  {
    "task": "TASK_DETECTION",
    "model_instance": "models/dummy-model/instances/v1.0",
    "task_outputs": [
      {
        "index": "01GB5T5ZK9W9C2VXMWWRYM8WPM",
        "detection": {
          "objects": [
            {
              "bounding_box": { "height": 0, "left": 0, "top": 99.084984, "width": 204.18988 },
              "category": "dog",
              "score": 0.980409
            },
            {
              "bounding_box": { "height": 242.36627, "left": 133.76924, "top": 195.17859, "width": 207.40651 },
              "category": "dog",
              "score": 0.9009272
            }
          ]
        }
      },
      {
        "index": "01GB5T5ZK9W9C2VXMWWRYM8WPN",
        "detection": {
          "objects": [
            {
              "bounding_box": { "height": 402.58002, "left": 0, "top": 99.084984, "width": 204.18988 },
              "category": "dog",
              "score": 0.980409
            },
            {
              "bounding_box": { "height": 242.36627, "left": 133.76924, "top": 195.17859, "width": 207.40651 },
              "category": "dog",
              "score": 0.9009272
            }
          ]
        }
      },
      {
        "index": "01GB5T5ZK9W9C2VXMWWRYM8WPO",
        "detection": {
          "objects": [
            {
              "bounding_box": { "height": 0, "left": 325.7926, "top": 99.084984, "width": 204.18988 },
              "category": "dog",
              "score": 0.980409
            },
            {
              "bounding_box": { "height": 242.36627, "left": 133.76924, "top": 195.17859, "width": 207.40651 },
              "category": "dog",
              "score": 0.9009272
            }
          ]
        }
      }
    ]
  },
  {
    "task": "TASK_DETECTION",
    "model_instance": "models/dummy-model/instances/v2.0",
    "task_outputs": [
      {
        "index": "01GB5T5ZK9W9C2VXMWWRYM8WPM",
        "detection": {
          "objects": [
            {
              "bounding_box": { "height": 0, "left": 0, "top": 99.084984, "width": 204.18988 },
              "category": "dog",
              "score": 0.980409
            },
            {
              "bounding_box": { "height": 242.36627, "left": 133.76924, "top": 195.17859, "width": 207.40651 },
              "category": "dog",
              "score": 0.9009272
            }
          ]
        }
      },
      {
        "index": "01GB5T5ZK9W9C2VXMWWRYM8WPN",
        "detection": {
          "objects": [
            {
              "bounding_box": { "height": 402.58002, "left": 0, "top": 99.084984, "width": 204.18988 },
              "category": "dog",
              "score": 0.980409
            },
            {
              "bounding_box": { "height": 242.36627, "left": 133.76924, "top": 195.17859, "width": 207.40651 },
              "category": "dog",
              "score": 0.9009272
            }
          ]
        }
      },
      {
        "index": "01GB5T5ZK9W9C2VXMWWRYM8WPO",
        "detection": {
          "objects": [
            {
              "bounding_box": { "height": 0, "left": 325.7926, "top": 99.084984, "width": 204.18988 },
              "category": "dog",
              "score": 0.980409
            },
            {
              "bounding_box": { "height": 242.36627, "left": 133.76924, "top": 195.17859, "width": 207.40651 },
              "category": "dog",
              "score": 0.9009272
            }
          ]
        }
      }
    ]
  }
]

export const detectionEmptyModelInstOutputs = [
  {
    "task": "TASK_DETECTION",
    "model_instance": "models/dummy-model/instances/v1.0",
    "task_outputs": [
      {
        "index": "01GB5T5ZK9W9C2VXMWWRYM8WPM",
        "detection": {
          "objects": []
        }
      },
    ]
  }
]

export const keypointModelInstOutputs = [
  {
    "task": "TASK_KEYPOINT",
    "model_instance": "models/dummy-model/instances/v1.0",
    "task_outputs": [
      {
        "index": "01GB5T5ZK9W9C2VXMWWRYM8WPT",
        "keypoint": {
          "objects": [
            {
              "keypoints": [{ "x": 10, "y": 100, "v": 0.6 }, { "x": 11, "y": 101, "v": 0.2 }],
              "score": 0.99
            },
            {
              "keypoints": [{ "x": 20, "y": 10, "v": 0.6 }, { "x": 12, "y": 120, "v": 0.7 }],
              "score": 0.99
            },
          ]
        }
      }
    ]
  }
]

export const ocrModelInstOutputs = [
  {
    "task": "TASK_OCR",
    "model_instance": "models/dummy-model/instances/v1.0",
    "task_outputs": [
      {
        "index": "01GB5T5ZK9W9C2VXMWWRYM8WPU",
        "ocr": {
          "objects": [
            {
              "bounding_box": { "height": 402.58002, "left": 0, "top": 99.084984, "width": 204.18988 },
              "text": "some text",
              "score": 0.99
            },
            {
              "bounding_box": { "height": 242.36627, "left": 133.76924, "top": 195.17859, "width": 207.40651 },
              "text": "some text",
              "score": 0.99
            },
          ],
        }
      }
    ]
  }
]

export const unspecifiedModelInstOutputs = [
  {
    "task": "TASK_UNSPECIFIED",
    "model_instance": "models/dummy-model/instances/v1.0",
    "task_outputs": [
      {
        "index": "01GB5T5ZK9W9C2VXMWWRYM8WPV",
        "unspecified": {
          "raw_outputs": [
            {
              "name": "some unspecified model output",
              "data_type": "INT8",
              "shape": [3, 3, 3],
              "data": [1, 2, 3, 4, 5, 6, 7]
            },
          ],
        }
      }
    ]
  }
]
