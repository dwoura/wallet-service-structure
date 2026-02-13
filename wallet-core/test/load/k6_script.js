import http from "k6/http";
import { check, sleep } from "k6";
import { randomString } from "https://jslib.k6.io/k6-utils/1.2.0/index.js";

// 1. 配置测试选项 (Options)
// 模拟 50 个并发用户 (Virtual Users)，持续 30 秒
export const options = {
  vus: 50,
  duration: "30s",
  thresholds: {
    http_req_duration: ["p(95)<500"], // 95% 的请求必须在 500ms 内完成
    http_req_failed: ["rate<0.01"], // 错误率必须低于 1%
  },
};

const BASE_URL = "http://localhost:8080/v1";

export default function () {
  // 2. 场景模拟: 用户注册 -> 登录 -> 查看余额

  // Step 1: 注册 (Register)
  const username = `user_${randomString(8)}`;
  const email = `${username}@example.com`;

  const registerPayload = JSON.stringify({
    username: username,
    password: "password123",
    email: email,
  });

  const params = {
    headers: {
      "Content-Type": "application/json",
    },
  };

  const resRegister = http.post(
    `${BASE_URL}/user/register`,
    registerPayload,
    params,
  );

  // 检查点 (Check): 确保注册成功
  check(resRegister, {
    "register status is 200": (r) => r.status === 200,
    "has user_id": (r) => JSON.parse(r.body).user_id !== undefined,
  });

  if (resRegister.status !== 200) {
    console.error(`Register failed: ${resRegister.body}`);
    return;
  }

  const userId = JSON.parse(resRegister.body).user_id;

  // Step 2: 模拟登录 (Login) - 虽然当前 Gateway 是 mock 实现，但流程要完整
  const loginPayload = JSON.stringify({
    email: email,
    password: "password123",
  });

  const resLogin = http.post(`${BASE_URL}/user/login`, loginPayload, params);
  check(resLogin, {
    "login status is 200": (r) => r.status === 200,
  });

  // Step 3: 查询余额 (Query Balance)
  // 模拟思考时间 (Think Time)
  sleep(1);

  const resBalance = http.get(
    `${BASE_URL}/wallet/balance?user_id=${userId}&currency=ETH`,
  );
  check(resBalance, {
    "balance status is 200": (r) => r.status === 200,
  });
}
