import React from 'react';
import { rouletteWheelOrder, getRouletteColor } from '../lib/roulette';
import './RouletteWheelSVG.css';

interface RouletteWheelSVGProps {
  rotation: number;
  spinning: boolean;
}

export const RouletteWheelSVG: React.FC<RouletteWheelSVGProps> = ({ rotation, spinning }) => {
  const count = rouletteWheelOrder.length;
  const anglePerSector = 360 / count;
  const cx = 120, cy = 120, r = 110, innerR = 42;

  const sectorPath = (index: number) => {
    const startAngle = (index * anglePerSector - 90) * (Math.PI / 180);
    const endAngle = ((index + 1) * anglePerSector - 90) * (Math.PI / 180);
    const x1 = cx + r * Math.cos(startAngle);
    const y1 = cy + r * Math.sin(startAngle);
    const x2 = cx + r * Math.cos(endAngle);
    const y2 = cy + r * Math.sin(endAngle);
    const ix1 = cx + innerR * Math.cos(startAngle);
    const iy1 = cy + innerR * Math.sin(startAngle);
    const ix2 = cx + innerR * Math.cos(endAngle);
    const iy2 = cy + innerR * Math.sin(endAngle);
    return `M${ix1},${iy1} L${x1},${y1} A${r},${r} 0 0,1 ${x2},${y2} L${ix2},${iy2} A${innerR},${innerR} 0 0,0 ${ix1},${iy1} Z`;
  };

  const labelPos = (index: number) => {
    const mid = (index + 0.5) * anglePerSector - 90;
    const lr = (r + innerR) / 2;
    return {
      x: cx + lr * Math.cos(mid * Math.PI / 180),
      y: cy + lr * Math.sin(mid * Math.PI / 180),
      angle: mid + 90,
    };
  };

  const getColorHex = (val: number) => {
    const color = getRouletteColor(val);
    if (color === 'green') return '#16a34a'; // Green 600
    if (color === 'red') return '#ef4444';   // Red 500
    return '#1d4ed8'; // Blue 700 (Black equivalent for Mudro)
  };

  return (
    <div className="roulette-wheel-svg">
      <svg
        width={240}
        height={240}
        viewBox="0 0 240 240"
        className="roulette-wheel-svg__svg"
        style={{
          transform: `rotate(${rotation}deg)`,
          transition: spinning
            ? 'none'
            : 'transform 4s cubic-bezier(0.17, 0.67, 0.12, 1)',
        }}
      >
        {/* Outer ring */}
        <circle cx={cx} cy={cy} r={r + 6} fill="#2a1a00" stroke="#f5c842" strokeWidth="2" />

        {/* Sectors */}
        {rouletteWheelOrder.map((val, i) => (
          <path 
            key={`${i}-${val}`} 
            d={sectorPath(i)} 
            fill={getColorHex(val)} 
            stroke="#111" 
            strokeWidth="0.5" 
          />
        ))}

        {/* Labels */}
        {rouletteWheelOrder.map((val, i) => {
          const pos = labelPos(i);
          return (
            <text
              key={`label-${i}-${val}`}
              x={pos.x}
              y={pos.y}
              textAnchor="middle"
              dominantBaseline="central"
              fontSize={count > 20 ? 5.5 : 8}
              fontWeight="700"
              fill="white"
              transform={`rotate(${pos.angle}, ${pos.x}, ${pos.y})`}
            >
              {val}
            </text>
          );
        })}

        {/* Inner hub */}
        <circle cx={cx} cy={cy} r={innerR} fill="#1a0f00" stroke="#f5c842" strokeWidth="2" />
        <circle cx={cx} cy={cy} r={innerR - 8} fill="#0d0d1a" />
        <text 
          x={cx} 
          y={cy} 
          textAnchor="middle" 
          dominantBaseline="central" 
          fontSize="14" 
          fill="#f5c842" 
          fontWeight="900"
          className="roulette-wheel-svg__hub-text"
        >
          🎰
        </text>
      </svg>

      {/* Ball pointer */}
      <div className="roulette-wheel-svg__pointer" />
    </div>
  );
};
