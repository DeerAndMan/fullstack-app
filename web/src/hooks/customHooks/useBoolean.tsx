import { useCallback, useState } from "react";

export interface BooleanType {
  value: boolean;
  onTrue: () => void;
  onFalse: () => void;
  onToggle: () => void;
  setValue: React.Dispatch<React.SetStateAction<boolean>>;
}

/**
 * 自定义Hook，创建一个包含布尔状态值和实用功能的对象
 *
 * @param {boolean} [defaultValue=false] - 布尔状态的初始值。
 */
export function useBoolean(defaultValue?: boolean): BooleanType {
  const [value, setValue] = useState(!!defaultValue);

  const onTrue = useCallback(() => {
    setValue(true);
  }, []);

  const onFalse = useCallback(() => {
    setValue(false);
  }, []);

  const onToggle = useCallback(() => {
    setValue(prev => !prev);
  }, []);

  return {
    value,
    onTrue,
    onFalse,
    onToggle,
    setValue,
  };
}

export default useBoolean;
