import sys
import subprocess
import shlex
import csv
import json
import yaml
import jsonschema
import tempfile
import os


def test_destination_csv(case: str):

    tmpCSV = tempfile.NamedTemporaryFile()

    subprocess.call(shlex.split(
        f'docker run -it --privileged --pid=host debian nsenter -t 1 -m -u -n -i sh -c "cat /var/lib/docker/volumes/airbyte/_data/test-{case}/_airbyte_raw_vdp.csv"'),
        stdout=open(tmpCSV.name, "w"))

    # The VDP protocol YAML file downloaded during image build time
    with open("/usr/local/vdp/vdp_protocol.yaml") as f:
        jsonSchema = json.loads(json.dumps(yaml.safe_load(f)))

    # read csv file
    with open(tmpCSV.name, encoding="utf-8") as csvf:
        # load csv file data using csv library's dictionary reader
        csvReader = csv.DictReader(csvf)

        if len(list(csvReader)) == 0:
            sys.exit(1)

        # convert each csv row into python dict
        for row in csvReader:
            try:
                jsonschema.validate(
                    instance=json.loads(row["_airbyte_data"]),
                    schema=jsonSchema,
                    cls=jsonschema.Draft7Validator,
                )
            except:
                sys.exit(1)


if __name__ == "__main__":

    test_cases = [
        "classification",
        "detection-empty-bounding-boxes",
        "detection-multi-models",
        "keypoint",
        "ocr",
        "unspecified"
    ]

    for case in test_cases:
        test_destination_csv(case)

    sys.exit(0)
