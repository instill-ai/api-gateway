import {check, group} from "k6";
import http from "k6/http";
import encoding from "k6/encoding";
// See https://github.com/szkiba/xk6-jose
import * as jwt from "k6/x/jose/jwt";
import * as jwk from "k6/x/jose/jwk";

const equals = (a, b) => a.length === b.length && a.every((v, i) => v === b[i]);

const issuer = "https://instill.tech/"; // set as HYDRA_ISSUER in .env.dev
const set = "hydra.jwt.access-token";

/*
 * Import a test user into Kratos
 */
export function importTestUser(kratosHost, testUserEmail) {
  let res;
  /*
   * Kratos - Create test user
   */
  group("Helpers - Kratos admin API: Create a test user", function () {
    res = http.request(
      "POST",
      `${kratosHost}/identities`,
      JSON.stringify({
        schema_id: "default",
        traits: {
          email: testUserEmail,
        },
      })
    );
    check(res, {
      [`POST /identities response status is 201`]: (r) => r.status === 201,
      "POST /identities response email check": (r) =>
        r.json("traits").email === testUserEmail,
      "POST /identities response id check": (r) => r.json("id") !== undefined,
    });
  });

  const testUser = res.json();

  return testUser;
}

/*
 * Generate a test user's token
 * the function imports a test user in Kratos and returns the test user and the test user's access token
 */
export function generateUserToken(
  kratosHost,
  hydraHost,
  audience,
  testUserEmail
) {
  const testUser = importTestUser(kratosHost, testUserEmail);
  let jwkset, priKid, pubKid, priKey, pubKey, userAccessToken;
  /*
   * Hydra - Get access token
   */
  {
    group("Helpers - Hydra admin API: Get an user access token", function () {
      jwkset = http.request("GET", `${hydraHost}/keys/${set}`).json("keys");

      check(jwkset, {
        [`Hydra: GET /keys/${set} has two keys`]: (j) => j.length === 2,
      });

      jwkset.forEach((jwk) => {
        if (jwk.kid.startsWith("private")) {
          priKid = jwk.kid;
        } else if (jwk.kid.startsWith("public")) {
          pubKid = jwk.kid;
        }
      });
      check(null, {
        "private jwk exists": (r) => priKid !== undefined,
        "public jwk exists": (r) => pubKid !== undefined,
      });

      let res = http
        .request("GET", `${hydraHost}/keys/${set}/${priKid}`)
        .json("keys");

      check(res, {
        [`Hydra: GET /keys/${set}/${priKid} fetch private key success`]: (r) =>
          r.length === 1,
      });
      priKey = jwk.parse(JSON.stringify(res[0]));

      res = http
        .request("GET", `${hydraHost}/keys/${set}/${pubKid}`)
        .json("keys");

      check(res, {
        [`Hydra: GET /keys/${set}/${priKid} fetch public key success`]: (r) =>
          r.length === 1,
      });
      pubKey = jwk.parse(JSON.stringify(res[0]));

      let header = {
        alg: "RS256",
        kid: pubKid,
      };

      const now = Date.now();
      const payload = {
        client_id: "0ce54f68-6867-47fb-92b9-30451e806e4b",
        sub: testUser.id,
        jti: "2bf06830-a91c-4a97-8e99-e46d3684e70f",
        iss: issuer,
        aud: [audience],
        ext: {},
        iat: Math.floor(now / 1000),
        exp: Math.floor(new Date(now + 24 * 60 * 60 * 1000).getTime() / 1000),
        scp: [
          "offline",
          "offline_access",
          "openid",
          "email",
          "profile",
          "pipelines.*",
          "models.*",
          "clients.*",
        ],
        username: testUserEmail.split("@")[0]
      };

      userAccessToken = jwt.sign(priKey, payload, header);

      let verifiedPayload = jwt.verify(userAccessToken, pubKey);
      check(null, {
        [`User access token issued successfully: ${userAccessToken}`]: (r) =>
          userAccessToken.length > 0,
        "User access token verified by public key successfully": (r) =>
          verifiedPayload.client_id === payload.client_id &&
          verifiedPayload.sub === payload.sub &&
          verifiedPayload.jti === payload.jti &&
          verifiedPayload.iss === payload.iss &&
          verifiedPayload.iat === payload.iat &&
          verifiedPayload.exp === payload.exp &&
          equals(verifiedPayload.aud, payload.aud) &&
          equals(verifiedPayload.scp, payload.scp) &&
          JSON.stringify(verifiedPayload.ext) === JSON.stringify(payload.ext),
      });
    });
  }

  return [testUser, userAccessToken];
}

/*
 * Generate a test client's token
 * the function creates a test client via mgmt-backend and returns the test client and the test client's access token
 */
// export function generateClientToken(userAccessToken, apiHost, audience) {
//   let testClient, clientAccessToken;

//   group(
//     "Helpers - Management API [with User token]: Create a client",
//     function () {
//       const clientName = "Test Client";
//       const clientDesc = "This is a test client.";
//       let res = http.request(
//         "POST",
//         `${apiHost}/clients`,
//         JSON.stringify({
//           name: clientName,
//           description: clientDesc,
//         }),
//         {
//           headers: genAuthHeader(userAccessToken),
//         }
//       );
//       check(res, {
//         "POST /clients response status is 201": (r) => r.status === 201,
//         "POST /clients response id check": (r) => r.json("id") !== undefined,
//         "POST /clients response secret check": (r) =>
//           r.json("secret") !== undefined,
//         "POST /clients response name check": (r) =>
//           r.json("name") === clientName,
//         "POST /clients response description check": (r) =>
//           r.json("description") === clientDesc,
//         "POST /clients response created_at check": (r) =>
//           r.json("created_at") !== undefined,
//         "POST /clients response update_at check": (r) =>
//           r.json("updated_at") !== undefined,
//         "POST /clients response kind check": (r) => r.json("kind") === "Client",
//         "POST /clients response self check": (r) =>
//           r.json("self") === `${apiHost}/clients/${r.json("id")}`,
//         "POST /clients response duration": (r) =>
//           r.json("duration") !== undefined,
//       });

//       testClient = res.json();
//     }
//   );

//   group("Helpers - Management API : Get a client token", function () {
//     const basicAuthHeader = encoding.b64encode(
//       `${testClient.id}:${testClient.secret}`
//     );
//     let res = http.request(
//       "POST",
//       `${apiHost}/oauth2/token?audience=${audience}&grant_type=client_credentials`,
//       null,
//       {
//         headers: {
//           "Content-Type": "application/json",
//           Authorization: `Basic ${basicAuthHeader}`,
//         },
//       }
//     );

//     check(res, {
//       [`POST /oauth2/token?audience=${audience}&grant_type=client_credentials response status is 200`]:
//         (r) => r.status === 200,
//       [`POST /oauth2/token?audience=${audience}&grant_type=client_credentials response kind`]:
//         (r) => r.json("kind") === "OAuth2JWT",
//       [`POST /oauth2/token?audience=${audience}&grant_type=client_credentials response access_token issued successfully: ${res.json(
//         "access_token"
//       )}`]: (r) => r.json("access_token") !== undefined,
//       [`POST /oauth2/token?audience=${audience}&grant_type=client_credentials response expires_in`]:
//         (r) => r.json("expires_in") !== undefined,
//       [`POST /oauth2/token?audience=${audience}&grant_type=client_credentials response expiry`]:
//         (r) => r.json("expiry") !== undefined,
//       [`POST /oauth2/token?audience=${audience}&grant_type=client_credentials response token_type`]:
//         (r) => r.json("token_type") === "Bearer",
//       [`POST /oauth2/token?audience=${audience}&grant_type=client_credentials response duration`]:
//         (r) => r.json("duration") !== undefined,
//     });

//     clientAccessToken = res.json("access_token");
//   });

//   return [testClient, clientAccessToken];
// }

/*
 * USE WITH CAUSION!!!
 * Delete a user in Kratos
 */
export function deleteUser(kratosHost, userID) {
  group("Helpers - Kratos admin API: Delete a user", function () {
    check(http.request("GET", `${kratosHost}/identities/${userID}`), {
      [`GET /identities/${userID} response status is 200`]: (r) =>
        r.status === 200,
    });

    check(http.request("DELETE", `${kratosHost}/identities/${userID}`), {
      [`DELETE /identities/${userID} response status is 204`]: (r) =>
        r.status === 204,
    });
  });
}

/*
 * USE WITH CAUSION!!!
 * Delete all clients of a user in Hydra
 */
export function deleteUserClients(hydraHost, userID) {
  const limit = 50;
  let offset = 0;
  group("Helpers - Hydra admin API: Delete all clients of a user", function () {
    while (true) {
      const clis = http
        .request(
          "GET",
          `${hydraHost}/clients?offset=${offset}&limit=${limit}&owner=${userID}`
        )
        .json();

      for (const cli of clis) {
        check(http.request("DELETE", `${hydraHost}/clients/${cli.client_id}`), {
          [`DELETE /clients/${cli.client_id} response status is 204`]: (r) =>
            r.status === 204,
        });
      }
      if (clis.length < limit) {
        break;
      }
      offset += limit;
    }
  });
}

export function genAuthHeader(accessToken, contentType) {
  return {
    "Content-Type": `${contentType}`,
    Authorization: `Bearer ${accessToken}`,
  };
}

export function genNoAuthHeader(contentType) {
  return {
    "Content-Type": `${contentType}`,
  };
}
