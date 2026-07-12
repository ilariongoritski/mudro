import { SlotMachine } from "@/components/slot/SlotMachine";
import { GameLobby } from "@/components/GameLobby";
import { HistoryPanel } from "@/components/HistoryPanel";

export default function Home() {
  return (
    <>
      <SlotMachine />
      <GameLobby />
      <HistoryPanel />
    </>
  );
}
