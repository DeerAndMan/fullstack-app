import z from "zod";

import request from "./request";

import type { Any } from "@/types/constants";
import type { CustomRequestConfig } from "./request";

type TypedRequestConfig<S extends z.ZodTypeAny = z.ZodTypeAny> = Omit<
  CustomRequestConfig,
  "headers" | "schema"
> & {
  schema: S; // 响应数据验证  schema
};

// 从 schema 推断 data 类型
type InferDataSchema<T extends z.ZodTypeAny> =
  T extends z.ZodArray<infer U> ? z.infer<U>[] : z.infer<T>;

// 推断完整的响应类型
type InferResponseSchema<T extends z.ZodTypeAny> = {
  data: InferDataSchema<T>;
  message: string;
};

type InferResponseSchemaCode<T extends z.ZodTypeAny> = InferResponseSchema<T> & { code: number };

// type PromiseResponse<T = Any> = Promise<ResponseData<T>>;

// 基础响应类型
export const BaseResponseSchema = z.object({
  code: z.number(),
  message: z.string(),
});

export type SchemaType = "base" | "table";
type DataSchema = Record<SchemaType, z.ZodTypeAny>;

const mergedSchema = (config: TypedRequestConfig): TypedRequestConfig => {
  const { schema, schemaType, ...rest } = config;

  const mergedMap: DataSchema = {
    base: schema,
    table: z.object({
      pageNumber: z.number(),
      pageSize: z.number(),
      totalPage: z.number(),
      totalCount: z.number(),
      list: z.array(schema),
    }),
  };

  return {
    ...rest,
    schema: BaseResponseSchema.extend({
      data: mergedMap[schemaType || "base"],
    }),
  };
};

/**
 * 类型化的请求函数，从 schema 自动推断返回类型
 * @param config 请求配置
 * @returns
 */
export const RequestSchema = async <T extends z.ZodTypeAny>(
  config: TypedRequestConfig<T>
): Promise<InferResponseSchema<T>> => {
  const finalConfig = mergedSchema(config);
  const res = (await request(finalConfig)) as InferResponseSchemaCode<T>;
  validateResponse(finalConfig.schema, res);
  return handleRequestCode<T>(res);
};

/**
 * GET 请求的类型化函数
 * @param url 请求地址
 * @param config 请求配置
 * @returns 返回类型
 */
export const RequestGet = async <T extends z.ZodTypeAny>(
  url: string,
  config: TypedRequestConfig<T>
): Promise<InferResponseSchema<T>> => {
  const finalConfig = mergedSchema(config);
  const res = await request.get(url, finalConfig);
  validateResponse(finalConfig.schema, res);
  return handleRequestCode<T>(res);
};

/**
 * POST 请求的类型化函数
 * @param url 请求地址
 * @param data 请求数据
 * @param config 请求配置
 * @returns 返回类型
 */
export const RequestPost = async <T extends z.ZodTypeAny>(
  url: string,
  data: Any,
  config: TypedRequestConfig<T>
): Promise<InferResponseSchema<T>> => {
  const finalConfig = mergedSchema(config);
  const res = await request.post(url, data, finalConfig);
  validateResponse(finalConfig.schema, res);
  return handleRequestCode<T>(res as InferResponseSchemaCode<T>);
};

/**
 * PUT 请求的类型化函数
 * @param url 请求地址
 * @param data 请求数据
 * @param config 请求配置
 * @returns 返回类型
 */
export const RequestPut = async <T extends z.ZodTypeAny>(
  url: string,
  data: Any,
  config: TypedRequestConfig<T>
): Promise<InferResponseSchema<T>> => {
  const finalConfig = mergedSchema(config);
  const res = await request.put(url, data, finalConfig);
  validateResponse(finalConfig.schema, res);
  return handleRequestCode<T>(res as InferResponseSchemaCode<T>);
};

/**
 * DELETE 请求的类型化函数
 * @param url 请求地址
 * @param config 请求配置
 * @returns 返回类型
 */
export const RequestDelete = async <T extends z.ZodTypeAny>(
  url: string,
  config: TypedRequestConfig<T>
): Promise<InferResponseSchema<T>> => {
  const finalConfig = mergedSchema(config);
  const res = await request.delete(url, finalConfig);
  validateResponse(finalConfig.schema, res);
  return handleRequestCode<T>(res as InferResponseSchemaCode<T>);
};

// 统一处理 请求Code
const handleRequestCode = <T extends z.ZodTypeAny>(res: InferResponseSchemaCode<T>) => {
  const { code, data, message } = res;

  if (code !== 0) {
    return Promise.reject(new Error(message));
  }

  return { data, message };
};

const validateResponse = (schema: z.ZodTypeAny, data: any) => {
  try {
    schema.parse(data);
  } catch (error) {
    if (error instanceof z.ZodError) {
      const errorMessages = error.issues
        .map((err) => `${err.path.join(".")}: ${err.message}`)
        .join(", ");
      console.warn(`响应数据验证失败: ${errorMessages}`);
    }
    throw error;
  }
};
