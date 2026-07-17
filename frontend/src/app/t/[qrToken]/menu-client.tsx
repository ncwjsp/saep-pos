"use client";

import { useMutation, useQuery } from "@tanstack/react-query";
import { useMemo } from "react";

import { Button } from "@/components/ui/button";
import {
  createOrder,
  fetchMenu,
  formatBaht,
  type MenuItem,
  type Order,
} from "@/lib/api";
import { cartCount, useCart } from "@/lib/cart";

export function MenuClient({ qrToken }: { qrToken: string }) {
  const menuQuery = useQuery({
    queryKey: ["menu", qrToken],
    queryFn: () => fetchMenu(qrToken),
  });

  const { quantities, add, remove, clear } = useCart();

  const submit = useMutation({
    mutationFn: () =>
      createOrder(
        qrToken,
        Object.entries(quantities).map(([menu_item_id, quantity]) => ({
          menu_item_id,
          quantity,
        })),
      ),
    onSuccess: () => clear(),
  });

  const byCategory = useMemo(() => {
    const groups = new Map<string, MenuItem[]>();
    for (const item of menuQuery.data ?? []) {
      const list = groups.get(item.category) ?? [];
      list.push(item);
      groups.set(item.category, list);
    }
    return groups;
  }, [menuQuery.data]);

  const totalSatang = useMemo(() => {
    if (!menuQuery.data) return 0;
    return menuQuery.data.reduce(
      (sum, item) => sum + item.price_satang * (quantities[item.id] ?? 0),
      0,
    );
  }, [menuQuery.data, quantities]);

  if (submit.isSuccess) {
    return (
      <OrderConfirmation
        order={submit.data}
        onOrderMore={() => submit.reset()}
      />
    );
  }

  return (
    <div className="mx-auto flex min-h-dvh w-full max-w-md flex-col">
      <header className="sticky top-0 z-10 border-b bg-background/95 px-4 py-3 backdrop-blur">
        <h1 className="text-lg font-semibold">แสบ POS</h1>
        <p className="text-sm text-muted-foreground">โต๊ะ {qrToken}</p>
      </header>

      <main className="flex-1 px-4 pb-32">
        {menuQuery.isPending && (
          <p className="py-10 text-center text-muted-foreground">
            กำลังโหลดเมนู…
          </p>
        )}
        {menuQuery.isError && (
          <p className="py-10 text-center text-destructive">
            โหลดเมนูไม่สำเร็จ: {menuQuery.error.message}
          </p>
        )}

        {[...byCategory.entries()].map(([category, items]) => (
          <section key={category} className="mt-6">
            <h2 className="mb-2 text-base font-semibold">{category}</h2>
            <ul className="divide-y rounded-xl border">
              {items.map((item) => (
                <MenuRow
                  key={item.id}
                  item={item}
                  quantity={quantities[item.id] ?? 0}
                  onAdd={() => add(item.id)}
                  onRemove={() => remove(item.id)}
                />
              ))}
            </ul>
          </section>
        ))}
      </main>

      {cartCount(quantities) > 0 && (
        <footer className="fixed inset-x-0 bottom-0 z-10 mx-auto w-full max-w-md border-t bg-background p-4">
          {submit.isError && (
            <p className="mb-2 text-sm text-destructive">
              สั่งไม่สำเร็จ: {submit.error.message}
            </p>
          )}
          <Button
            className="h-14 w-full text-base"
            disabled={submit.isPending}
            onClick={() => submit.mutate()}
          >
            {submit.isPending
              ? "กำลังส่ง…"
              : `สั่งอาหาร ${cartCount(quantities)} รายการ · ${formatBaht(totalSatang)}`}
          </Button>
        </footer>
      )}
    </div>
  );
}

function MenuRow({
  item,
  quantity,
  onAdd,
  onRemove,
}: {
  item: MenuItem;
  quantity: number;
  onAdd: () => void;
  onRemove: () => void;
}) {
  return (
    <li className="flex items-center gap-3 p-3">
      <div className="min-w-0 flex-1">
        <p className="font-medium">{item.name}</p>
        <p className="truncate text-sm text-muted-foreground">{item.name_en}</p>
        <p className="mt-1 text-sm font-semibold">
          {formatBaht(item.price_satang)}
        </p>
      </div>
      {quantity === 0 ? (
        <Button
          variant="outline"
          className="h-11 px-5"
          onClick={onAdd}
          aria-label={`เพิ่ม ${item.name}`}
        >
          เพิ่ม
        </Button>
      ) : (
        <div className="flex items-center gap-1">
          <Button
            variant="outline"
            className="size-11 text-lg"
            onClick={onRemove}
            aria-label={`ลด ${item.name}`}
          >
            −
          </Button>
          <span className="w-8 text-center text-base font-semibold">
            {quantity}
          </span>
          <Button
            variant="outline"
            className="size-11 text-lg"
            onClick={onAdd}
            aria-label={`เพิ่ม ${item.name}`}
          >
            +
          </Button>
        </div>
      )}
    </li>
  );
}

function OrderConfirmation({
  order,
  onOrderMore,
}: {
  order: Order;
  onOrderMore: () => void;
}) {
  return (
    <div className="mx-auto flex min-h-dvh w-full max-w-md flex-col items-center justify-center gap-4 px-6 text-center">
      <div className="text-5xl">✅</div>
      <h1 className="text-xl font-semibold">สั่งอาหารเรียบร้อย</h1>
      <p className="text-muted-foreground">
        ออเดอร์ #{order.id} · ส่งเข้าครัวแล้ว
      </p>
      <ul className="w-full rounded-xl border text-left">
        {order.items.map((it, i) => (
          <li key={i} className="flex justify-between border-b p-3 last:border-b-0">
            <span>
              {it.name} × {it.quantity}
            </span>
            <span className="font-medium">
              {formatBaht(it.price_satang * it.quantity)}
            </span>
          </li>
        ))}
        <li className="flex justify-between p-3 font-semibold">
          <span>รวม</span>
          <span>{formatBaht(order.total_satang)}</span>
        </li>
      </ul>
      <Button className="h-12 w-full" onClick={onOrderMore}>
        สั่งเพิ่ม
      </Button>
    </div>
  );
}
