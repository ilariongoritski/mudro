import React, { useMemo } from 'react';
import { rouletteWheelOrder, getRouletteColor } from '../lib/roulette';
import './RouletteWheelSVG.css';

interface RouletteWheelSVGProps {
  rotation: number;
}

export const RouletteWheelSVG: React.FC<RouletteWheelSVGProps> = ({ rotation }) => {
  const count = rouletteWheelOrder.length;
  const anglePerSector = 360 / count;
  const cx = 120, cy = 120, r = 110, innerR = 42;

  const sectors = useMemo(() => {
    return rouletteWheelOrder.map((val, index) => {
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

      const mid = (index + 0.5) * anglePerSector - 90;
      const lr = (r + innerR) / 2;
      
      const color = getRouletteColor(val);
      let fill = '#1d4ed8'; // Blue 700 (Black equivalent for Mudro)
      if (color === 'green') fill = '#16a34a'; // Green 600
      if (color === 'red') fill = '#ef4444';   // Red 500

      return {
        val,
        path: `M${ix1},${iy1} L${x1},${y1} A${r},${r} 0 0,1 ${x2},${y2} L${ix2},${iy2} A${innerR},${innerR} 0 0,0 ${ix1},${iy1} Z`,
        fill,
        labelPos: {
          x: cx + lr * Math.cos(mid * Math.PI / 180),
          y: cy + lr * Math.sin(mid * Math.PI / 180),
          angle: mid + 90,
        }
      };
    });
  }, [anglePerSector, cx, cy, r, innerR]);

  return (
    <div className="roulette-wheel-svg">
      <svg
        width={240}
        height={240}
        viewBox="0 0 240 240"
        className="roulette-wheel-svg__svg"
        style={{
          transform: `rotate(${rotation}deg)`,
          transition: 'transform 4s cubic-bezier(0.12, 0.8, 0.33, 1)',
        }}
      >
        {/* Outer ring */}
        <circle cx={cx} cy={cy} r={r + 6} fill="#2a1a00" stroke="#f5c842" strokeWidth="2" />

        {/* Sectors */}
        {sectors.map((s, i) => (
          <path 
            key={`${i}-${s.val}`} 
            d={s.path} 
            fill={s.fill} 
            stroke="#111" 
            strokeWidth="0.5" 
          />
        ))}

        {/* Labels */}
        {sectors.map((s, i) => (
          <text
            key={`label-${i}-${s.val}`}
            x={s.labelPos.x}
            y={s.labelPos.y}
            textAnchor="middle"
            dominantBaseline="central"
            fontSize={count > 20 ? 5.5 : 8}
            fontWeight="700"
            fill="white"
            transform={`rotate(${s.labelPos.angle}, ${s.labelPos.x}, ${s.labelPos.y})`}
          >
            {s.val}
          </text>
        ))}

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
