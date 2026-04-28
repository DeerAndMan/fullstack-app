import { z } from "zod";
import axios from "axios";
import axiosRetry from "axios-retry";
import { message } from "antd";

import { useAuthStore } from "@/stores/auth";
import { setAuthCookie } from "@/utils/cookie";

import type { AxiosError, AxiosResponse, InternalAxiosRequestConfig } from "axios";
import type { Any } from "@/types/constants";
import type { SchemaType } from "./request-schema";

type RequestConfig = {
  noToken?: boolean;
  errorMsg?: boolean | string;
  successMsg?: boolean | string;
  saltLength?: number;
  needRetry?: boolean;
  schema?: z.ZodSchema;
  schemaType?: SchemaType;
};

export type CustomRequestConfig = RequestConfig & InternalAxiosRequestConfig;
export type PartialCustomRequestConfig = Partial<CustomRequestConfig>;

const statusList = [401, 403];
const tokenErrList = [401, 10002, 10003];

export const BASE_API = import.meta.env.VITE_WEB_BASE_URL;

export const request = axios.create({
  baseURL: BASE_API,
  timeout: 300000,
  headers: { Accept: "application/json", "Content-type": "application/json" },
});

export type ResponseData<T = Any> = {
  data: T;
  code: number;
  msg: string;
  message: string;
};

axiosRetry(request, {
  retries: 3,
  retryDelay: (retryCount) => {
    return retryCount * 1000;
  },
  retryCondition: (error) => {
    const config = error.config as CustomRequestConfig;
    return !!config.needRetry && error.response?.status === 500;
  },
});

request.interceptors.request.use((config: CustomRequestConfig) => {
  if (!config.noToken) {
    const { token } = useAuthStore.getState();
    if (token) {
      config.headers.Authorization = token;
    }
  }

  if (config.errorMsg === undefined) {
    config.errorMsg = true;
  }

  if (config.saltLength) {
    config.headers.SaltLength = config.saltLength;
  }

  return config;
});

export type PromiseResponseData<T = Any> = Promise<ResponseData<T>>;

request.interceptors.response.use(
  async (res: AxiosResponse) => {
    const config = res.config as CustomRequestConfig;
    const data = res.data as ResponseData;

    const newToken = res.headers["x-new-token"];
    if (newToken) {
      useAuthStore.getState().setToken(newToken);
      setAuthCookie(newToken);
    }

    let errorMsg, successMsg;
    if (config) {
      errorMsg = config.timeoutErrorMessage;
      successMsg = config.successMsg;
    }

    if (res.data instanceof Blob) {
      return res;
    }

    if (!(data instanceof Blob) && res.data.code === 0) {
      if (successMsg) {
        if (typeof successMsg === "string") {
          message.success(successMsg);
        } else {
          message.success(res.data.message ?? res.data.msg);
        }
      }
    } else if (!(data instanceof Blob) && res.data.code !== 0 && errorMsg) {
      if (typeof errorMsg === "string") {
        message.error(errorMsg);
      } else {
        message.error(res?.data?.message ?? res?.data?.msg ?? "出错啦");
      }
    }

    if (tokenErrList.includes(res.data.code)) {
      useAuthStore.getState().clearToken();
      return Promise.reject(new Error(res.data.message ?? res.data.msg));
    }

    if (res.data.code !== 0) {
      return Promise.reject(res.data);
    }

    return res.data;
  },
  (err: AxiosError) => {
    message.error(err.message);

    const { status, response } = err;
    const data = response?.data as ResponseData;

    if (status && statusList.includes(status) && tokenErrList.includes(data?.code)) {
      useAuthStore.getState().clearToken();
      return Promise.reject(new Error(data.msg));
    }

    return Promise.reject(err?.response?.data || "出错啦");
  },
);

export default request;
