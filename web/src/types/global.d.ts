import type { z } from "zod";
import type { Any } from "@/types/constants";
import type { PartialCustomRequestConfig } from "@/api";
import type { ResponseData } from "@/api";
import type { SchemaType } from "@/api/request-schema";

declare module "axios" {
  export interface AxiosRequestConfig {
    // 接口报错是否需要toast提示。若为true则提示，若为字符串则为自定义的提示内容
    errorMsg?: boolean | string;
    // 接口成功是否需要toast提示。若为true则提示，若为字符串则为自定义的提示内容
    successMsg?: boolean | string;
    // 接口是否不需要token（默认都会加上token）
    noToken?: boolean;
    // 返回接口原内容
    returnOrigin?: boolean;
    // 盐长度
    saltLength?: number;
    // 控制是否需要重试
    needRetry?: boolean;
    // 响应数据验证 schema
    schema?: z.ZodSchema;
    // 响应数据验证类型
    schemaType?: SchemaType;

    //   // [自定义属性声明]
    //   // 接口报错是否需要toast提示。若为true则提示，若为字符串则为自定义的提示内容
    //   errorMsg?: boolean | string;
    //   // 接口成功是否需要toast提示。若为true则提示，若为字符串则为自定义的提示内容
    //   successMsg?: boolean | string;
    //   // 接口是否不需要token（默认都会加上token）
    //   noToken?: boolean;
    //   // 返回接口原内容
    //   returnOrigin?: boolean;
    // }
    // interface ResponseProps<T = any> {
    //   data: T;
    //   code: number;
    //   msg: string;
    // }
    // export interface AxiosResponse<T = any, D = any> {
    //   data: ResponseProps<T>;
    //   status: number;
    //   statusText: string;
    //   headers: AxiosResponseHeaders;
    //   config: AxiosRequestConfig<D>;
    //   request?: any;
  }
  // 修正拦截器返回数据类型
  export interface AxiosInstance {
    get<T = Any>(url: string, config?: PartialCustomRequestConfig): Promise<ResponseData<T>>;
    post<T = Any>(
      url: string,
      data?: Any,
      config?: PartialCustomRequestConfig
    ): Promise<ResponseData<T>>;
    put<T = Any>(
      url: string,
      data?: Any,
      config?: PartialCustomRequestConfig
    ): Promise<ResponseData<T>>;
    delete<T = Any>(url: string, config?: PartialCustomRequestConfig): Promise<ResponseData<T>>;
  }
}

/// <reference types="@types/sigmajs" />
