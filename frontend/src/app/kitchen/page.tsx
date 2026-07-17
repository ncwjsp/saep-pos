"use client";

import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect, useState } from "react";

import { Badge } from "@/components/ui/badge";
import {
  API_URL,
  fetchKitchenOrders,
  formatBaht,
  type Order,
} from "@/lib/api";

const statusLabel: Record<Order["status"], string> = {
  pending: "รอทำ",
  preparing: "กำลังทำ",
  served: "เสิร์ฟแล้ว",
  cancelled: "ยกเลิก",
};

export default function KitchenPage() {
  const queryClient = useQueryClient();
  const [connected, setConnected] = useState(false);

  const ordersQuery = useQuery({
    queryKey: ["kitchen-orders"],
    queryFn: fetchKitchenOrders,
  });

  useEffect(() => {
    const es = new EventSource(`${API_URL}/kitchen/stream`);
    es.onopen = () => {
      setConnected(true);
      // Refetch on (re)connect to pick up orders missed while disconnected.
      queryClient.invalidateQueries({ queryKey: ["kitchen-orders"] });
    };
    // EventSource reconnects on its own; just reflect the state.
    es.onerror = () => setConnected(false);
    es.addEventListener("order_created", (e) => {
      const order = JSON.parse((e as MessageEvent).data) as Order;
      queryClient.setQueryData<Order[]>(["kitchen-orders"], (old) => {
        if (old?.some((o) => o.id === order.id)) return old;
        return [...(old ?? []), order];
      });
    });
    return () => es.close();
  }, [queryClient]);

  const orders = [...(ordersQuery.data ?? [])].reverse(); // newest first

  return (
    <div className="min-h-dvh w-full p-6">
      <header className="mb-6 flex items-center justify-between">
        <h1 className="text-2xl font-semibold">ครัว · แสบ POS</h1>
        <Badge variant={connected ? "secondary" : "destructive"}>
          {connected ? "● สดอยู่" : "○ กำลังเชื่อมต่อใหม่…"}
        </Badge>
      </header>

      {ordersQuery.isPending && (
        <p className="py-10 text-center text-muted-foreground">
          กำลังโหลดออเดอร์…
        </p>
      )}
      {ordersQuery.isError && (
        <p className="py-10 text-center text-destructive">
          โหลดออเดอร์ไม่สำเร็จ: {ordersQuery.error.message}
        </p>
      )}
      {ordersQuery.isSuccess && orders.length === 0 && (
        <p className="py-10 text-center text-muted-foreground">
          ยังไม่มีออเดอร์
        </p>
      )}

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        {orders.map((order) => (
          <OrderCard key={order.id} order={order} />
        ))}
      </div>
    </div>
  );
}

function OrderCard({ order }: { order: Order }) {
  const time = new Date(order.created_at).toLocaleTimeString("th-TH", {
    hour: "2-digit",
    minute: "2-digit",
  });

  return (
    <div className="rounded-xl border bg-card p-4 shadow-sm">
      <div className="mb-3 flex items-center justify-between">
        <span className="text-lg font-semibold">#{order.id}</span>
        <div className="flex items-center gap-2">
          <span className="text-sm text-muted-foreground">{time}</span>
          <Badge
            variant={order.status === "pending" ? "default" : "secondary"}
          >
            {statusLabel[order.status]}
          </Badge>
        </div>
      </div>
      <ul className="space-y-1">
        {order.items.map((it, i) => (
          <li key={i} className="flex justify-between gap-2">
            <span>
              {it.name}{" "}
              <span className="font-semibold">× {it.quantity}</span>
              {it.note && (
                <span className="block text-sm text-muted-foreground">
                  หมายเหตุ: {it.note}
                </span>
              )}
            </span>
          </li>
        ))}
      </ul>
      <div className="mt-3 border-t pt-2 text-right text-sm text-muted-foreground">
        รวม {formatBaht(order.total_satang)}
      </div>
    </div>
  );
}
