import { toast as reactToast } from "react-toastify";

import type { ToastContent, ToastOptions, Id } from "react-toastify";

type ToastFun = <TData = unknown>(
  content: ToastContent<TData>,
  options?: ToastOptions<TData>
) => Id;

interface Toast {
  info: ToastFun;
  warning: ToastFun;
  success: ToastFun;
  error: ToastFun;
}

const defaultOptions = {
  hideProgressBar: true,
  style: { borderRadius: "8px", color: "#000" },
};

/**
 * 提示方法集
 */
export const toast: Toast = {
  success: (content, options) =>
    reactToast.success(content, {
      ...defaultOptions,
      ...options,
    }),
  error: (content, options) => reactToast.error(content, options),
  warning: (content, options) => reactToast.warning(content, options),
  info: (content, options) => reactToast.info(content, options),
};
