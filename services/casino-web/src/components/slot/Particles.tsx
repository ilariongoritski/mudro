"use client";

import { useEffect, useRef } from "react";
import { useSlot } from "@/lib/slot/store";
import type { WinTier } from "@/lib/slot/config";

interface Candy {
  x: number;
  y: number;
  vx: number;
  vy: number;
  r: number;
  rot: number;
  vrot: number;
  life: number;
  maxLife: number;
  emoji: string;
}

const EMOJIS = ["🍭", "🍬", "🍫", "🍩", "🧁", "🍬", "🍭", "💎", "⭐", "💜", "💛"];

export function Particles() {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const itemsRef = useRef<Candy[]>([]);
  const rafRef = useRef<number | null>(null);
  const lastSpinKeyRef = useRef(0);

  const winTier = useSlot((s) => s.winTier);
  const spinKey = useSlot((s) => s.spinKey);
  const inFreeSpins = useSlot((s) => s.inFreeSpins);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const resize = () => {
      const dpr = Math.min(2, window.devicePixelRatio || 1);
      canvas.width = window.innerWidth * dpr;
      canvas.height = window.innerHeight * dpr;
      canvas.style.width = window.innerWidth + "px";
      canvas.style.height = window.innerHeight + "px";
      const ctx = canvas.getContext("2d");
      if (ctx) ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    };
    resize();
    window.addEventListener("resize", resize);
    return () => window.removeEventListener("resize", resize);
  }, []);

  useEffect(() => {
    lastSpinKeyRef.current = spinKey;
    if (winTier === "none" || winTier === "normal") return;
    // Trimmed counts: subtler, less excessive.
    const count =
      winTier === "epic" ? 48 : winTier === "mega" ? 32 : winTier === "big" ? 18 : 0;
    if (count === 0) return;
    const w = window.innerWidth;
    const cx = w / 2;
    const cy = window.innerHeight * 0.42;
    const items = itemsRef.current;
    for (let i = 0; i < count; i++) {
      const angle = -Math.PI / 2 + (Math.random() - 0.5) * Math.PI * 1.3;
      const speed = 4 + Math.random() * 6;
      items.push({
        x: cx + (Math.random() - 0.5) * 80,
        y: cy + (Math.random() - 0.5) * 30,
        vx: Math.cos(angle) * speed,
        vy: Math.sin(angle) * speed - 2,
        r: 12 + Math.random() * 8,
        rot: Math.random() * Math.PI,
        vrot: (Math.random() - 0.5) * 0.22,
        life: 0,
        maxLife: 100 + Math.random() * 60,
        emoji: EMOJIS[(Math.random() * EMOJIS.length) | 0],
      });
    }
  }, [winTier, spinKey]);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    const render = () => {
      ctx.clearRect(0, 0, window.innerWidth, window.innerHeight);
      const items = itemsRef.current;
      for (let i = items.length - 1; i >= 0; i--) {
        const c = items[i];
        c.life++;
        c.vy += 0.35;
        c.vx *= 0.996;
        c.x += c.vx;
        c.y += c.vy;
        c.rot += c.vrot;
        if (c.life > c.maxLife || c.y > window.innerHeight + 50) {
          items.splice(i, 1);
          continue;
        }
        const alpha = Math.max(0, 1 - c.life / c.maxLife);
        ctx.save();
        ctx.globalAlpha = alpha;
        ctx.translate(c.x, c.y);
        ctx.rotate(c.rot);
        ctx.font = `${c.r}px serif`;
        ctx.textAlign = "center";
        ctx.textBaseline = "middle";
        ctx.shadowColor = "rgba(0,0,0,.4)";
        ctx.shadowBlur = 4;
        ctx.fillText(c.emoji, 0, 0);
        ctx.restore();
      }
      rafRef.current = requestAnimationFrame(render);
    };
    rafRef.current = requestAnimationFrame(render);
    return () => {
      if (rafRef.current) cancelAnimationFrame(rafRef.current);
    };
  }, []);

  return (
    <canvas
      ref={canvasRef}
      className="fixed inset-0 pointer-events-none z-50"
      aria-hidden
      style={{ opacity: inFreeSpins ? 1 : 1 }}
    />
  );
}
