"use client";

import dynamic from "next/dynamic";

const SlotMachine = dynamic(() => import("@/components/slot/SlotMachine"), {
  ssr: false,
  loading: () => <div className="min-h-screen flex items-center justify-center text-white" />,
});

export default function Home() {
  return <SlotMachine />;
}