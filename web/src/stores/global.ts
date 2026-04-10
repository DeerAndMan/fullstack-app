import { create } from "zustand";

import type { MessageInstance } from "antd/lib/message/interface";

interface GlobalState {
  messageApi: MessageInstance | null;
  setMessageApi: (api: MessageInstance) => void;
}

export const useGlobalStore = create<GlobalState>()((set) => ({
  messageApi: null,
  setMessageApi: (api: MessageInstance) => set({ messageApi: api }),
}));
