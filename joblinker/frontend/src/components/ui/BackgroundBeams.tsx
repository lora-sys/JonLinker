'use client';

import { useEffect, useState } from 'react';
import { motion, useReducedMotion } from 'framer-motion';

interface Beam {
  x: number;
  delay: number;
  duration: number;
  height: number;
}

export function BackgroundBeams() {
  const [beams, setBeams] = useState<Beam[]>([]);
  const prefersReducedMotion = useReducedMotion();

  useEffect(() => {
    const newBeams = Array.from({ length: prefersReducedMotion ? 2 : 6 }, () => ({
      x: Math.random() * 100,
      delay: Math.random() * 5,
      duration: 10 + Math.random() * 10,
      height: 100 + Math.random() * 200,
    }));
    setBeams(newBeams);
  }, [prefersReducedMotion]);

  if (prefersReducedMotion) {
    return null; // Skip animation entirely
  }

  return (
    <div className="absolute inset-0 overflow-hidden pointer-events-none">
      {beams.map((beam, i) => (
        <motion.div
          key={i}
          className="absolute w-[1px] bg-gradient-to-b from-transparent via-blue-500/20 to-transparent"
          style={{ left: `${beam.x}%`, height: beam.height }}
          initial={{ y: '-100%', opacity: 0 }}
          animate={{ y: '200%', opacity: [0, 1, 1, 0] }}
          transition={{
            duration: beam.duration,
            delay: beam.delay,
            repeat: Infinity,
            ease: 'linear',
          }}
        />
      ))}
    </div>
  );
}
