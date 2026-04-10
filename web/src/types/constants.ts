// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type Any = any;

export interface CallbackFunction<T = Any> {
  successCb?: (data: T) => void;
  errorCb?: ({ msg }: { msg: string }) => void;
  finallyCb?: VoidFunction;
}
