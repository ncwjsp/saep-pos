import { create } from "zustand";

// Cart holds quantities keyed by menu_item_id. Prices always come from the
// menu data at render time; the server re-snapshots them on submit.
type CartState = {
  quantities: Record<string, number>;
  add: (menuItemId: string) => void;
  remove: (menuItemId: string) => void;
  clear: () => void;
};

export const useCart = create<CartState>((set) => ({
  quantities: {},
  add: (id) =>
    set((s) => ({
      quantities: { ...s.quantities, [id]: (s.quantities[id] ?? 0) + 1 },
    })),
  remove: (id) =>
    set((s) => {
      const next = { ...s.quantities };
      const q = (next[id] ?? 0) - 1;
      if (q <= 0) {
        delete next[id];
      } else {
        next[id] = q;
      }
      return { quantities: next };
    }),
  clear: () => set({ quantities: {} }),
}));

export function cartCount(quantities: Record<string, number>): number {
  return Object.values(quantities).reduce((sum, q) => sum + q, 0);
}
