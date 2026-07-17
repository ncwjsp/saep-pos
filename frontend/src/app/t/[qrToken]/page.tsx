import { MenuClient } from "./menu-client";

export default async function TablePage({ params }: PageProps<"/t/[qrToken]">) {
  const { qrToken } = await params;
  return <MenuClient qrToken={qrToken} />;
}
