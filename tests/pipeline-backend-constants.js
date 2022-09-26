import encoding from "k6/encoding";

const fooBarImg = open(`${__ENV.TEST_FOLDER_ABS_PATH}/cubo.jpg`, "b");

export const fooBarDetectionModel = {
  name: "m_3v2Yq6ocICEq0LxDdt8dBtl92Yl3QeWA",
  version: 1,
};
export const instillPublicClassficationModel = {
  name: "c64j7td9481af4asqb9g",
  version: 1,
};
export const instillPublicCocoDetectionModel = {
  name: "c64j81t9481afcrn2o60",
  version: 1,
};
export const unmarshallModel = { name: "c64j7td9481af4asqb9g", version: "1" };

export const classificationRecipe = {
  recipe: {
    source: {
      type: "direct",
    },
    model: [instillPublicClassficationModel],
    destination: {
      type: "direct",
    },
  },
};

export const cocoDetectionRecipe = {
  recipe: {
    source: {
      type: "direct",
    },
    model: [instillPublicCocoDetectionModel],
    destination: {
      type: "direct",
    },
  },
};

export const fooBarDetectionRecipe = {
  recipe: {
    source: {
      type: "direct",
    },
    model: [fooBarDetectionModel],
    destination: {
      type: "direct",
    },
  },
};

export const unmarshallRecipe = {
  recipe: {
    source: {
      type: "direct",
    },
    model: [unmarshallModel],
    destination: {
      type: "direct",
    },
  },
};

export const triggerPipelineJSONUrl = {
  contents: [
    {
      url: "https://artifacts.instill.tech/dog.jpg",
    },
  ],
};

export const triggerPipelineJSONBase64 = {
  contents: [
    {
      base64:
        "iVBORw0KGgoAAAANSUhEUgAAABAAAAAPBAMAAAAfXVIcAAAAD1BMVEV63/39//w5TVIZFhXDjXbHNiz1AAAARUlEQVR4nJTJUQ3AIAwG4aMzsBYD9FdQEfjXtJBiYPf0JYeHJKUTC8CShYNjaMwas4xoJNHrgKdo7H0x3ovTP3wBAAD//9u1Bcrd6KY0AAAAAElFTkSuQmCC",
    },
  ],
};

export const triggerPipelineJSONCuboBase64 = {
  contents: [
    {
      base64: encoding.b64encode(fooBarImg, "b"),
    },
  ],
};
