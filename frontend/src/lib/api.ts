const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export type MenuItem = {
  id: string;
  name: string;
  name_en: string;
  price_satang: number;
  category: string;
};

export type OrderItemInput = {
  menu_item_id: string;
  quantity: number;
  note?: string;
};

export type Order = {
  id: string;
  status: "pending" | "preparing" | "served" | "cancelled";
  items: {
    menu_item_id: string;
    name: string;
    price_satang: number;
    quantity: number;
    note?: string;
  }[];
  total_satang: number;
  created_at: string;
};

async function parseError(res: Response): Promise<string> {
  try {
    const body = (await res.json()) as { error?: string };
    if (body.error) return body.error;
  } catch {
    // non-JSON error body
  }
  return `request failed (${res.status})`;
}

export async function fetchMenu(qrToken: string): Promise<MenuItem[]> {
  const res = await fetch(`${API_URL}/t/${qrToken}/menu`);
  if (!res.ok) throw new Error(await parseError(res));
  const body = (await res.json()) as { items: MenuItem[] };
  return body.items;
}

export async function createOrder(
  qrToken: string,
  items: OrderItemInput[],
): Promise<Order> {
  const res = await fetch(`${API_URL}/t/${qrToken}/orders`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ items }),
  });
  if (!res.ok) throw new Error(await parseError(res));
  return (await res.json()) as Order;
}

// formatBaht renders satang as baht using integer math only (no float money).
export function formatBaht(satang: number): string {
  const whole = Math.trunc(satang / 100);
  const frac = Math.abs(satang % 100);
  const wholeStr = whole.toLocaleString("th-TH");
  return frac === 0
    ? `฿${wholeStr}`
    : `฿${wholeStr}.${String(frac).padStart(2, "0")}`;
}
