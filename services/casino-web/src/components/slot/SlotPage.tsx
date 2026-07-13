"use client";

import dynamic from "next/dynamic";

const SlotMachineDynamic = dynamic(() => import("./SlotMachine"), {
  ssr: false,
  loading: () => <div className="min-h-screen flex items-center justify-center text-white" />,
});

export default function SlotPage() {
  return <SlotMachineDynamic />;
}