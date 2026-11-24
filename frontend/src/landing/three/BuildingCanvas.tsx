import { useEffect, useRef, useState } from 'react';
import { Canvas, useFrame, useThree } from '@react-three/fiber';
import { RoundedBox } from '@react-three/drei';
import type { Group, Mesh } from 'three';

type SceneColors = {
  body: string;
  floor: string;
  panel: string;
  selected: string;
  hovered: string;
  keyLight: string;
  fillLight: string;
};

const fallbackColors: SceneColors = {
  body: '#dce5f2',
  floor: '#c7d3e4',
  panel: '#6878d8',
  selected: '#5365f5',
  hovered: '#8b7cf6',
  keyLight: '#fffaf0',
  fillLight: '#dfe7ff',
};

function readToken(name: string, fallback: string) {
  if (typeof document === 'undefined') return fallback;
  return getComputedStyle(document.querySelector('.lp-root') ?? document.documentElement).getPropertyValue(name).trim() || fallback;
}

function useSceneColors() {
  const [colors, setColors] = useState(fallbackColors);
  useEffect(() => {
    setColors({
      body: readToken('--lp-building-body', fallbackColors.body),
      floor: readToken('--lp-building-floor', fallbackColors.floor),
      panel: readToken('--lp-building-panel', fallbackColors.panel),
      selected: readToken('--lp-primary', fallbackColors.selected),
      hovered: readToken('--lp-accent', fallbackColors.hovered),
      keyLight: readToken('--lp-building-key-light', fallbackColors.keyLight),
      fillLight: readToken('--lp-building-fill-light', fallbackColors.fillLight),
    });
  }, []);
  return colors;
}

function CameraParallax() {
  const { camera } = useThree();
  useFrame(({ pointer }: any) => {
    camera.position.x += (pointer.x * 0.22 + 4.8 - camera.position.x) * 0.025;
    camera.position.y += (pointer.y * 0.14 + 3.5 - camera.position.y) * 0.025;
    camera.lookAt(0, 1.45, 0);
  });
  return null;
}

function WindowUnit({
  position,
  selected,
  hovered,
  onSelect,
  onHover,
  colors,
}: {
  position: [number, number, number];
  selected: boolean;
  hovered: boolean;
  onSelect: () => void;
  onHover: (value: boolean) => void;
  colors: SceneColors;
}) {
  const mesh = useRef<Mesh>(null);
  useFrame(() => {
    if (mesh.current) {
      const target = selected ? 1.08 : hovered ? 1.04 : 1;
      mesh.current.scale.x += (target - mesh.current.scale.x) * 0.12;
      mesh.current.scale.y += (target - mesh.current.scale.y) * 0.12;
    }
  });

  return (
    <group position={position}>
      <mesh position={[0, 0, -0.035]}>
        <boxGeometry args={[0.52, 0.38, 0.06]} />
        <meshStandardMaterial color="#f5f8ff" roughness={0.35} />
      </mesh>
      <mesh
        ref={mesh}
        position={[0, 0, 0.015]}
        onClick={onSelect}
        onPointerOver={(event: any) => {
          event.stopPropagation();
          onHover(true);
        }}
        onPointerOut={() => onHover(false)}
      >
        <boxGeometry args={[0.39, 0.26, 0.075]} />
        <meshStandardMaterial
          color={selected ? colors.selected : hovered ? colors.hovered : colors.panel}
          emissive={selected ? colors.selected : '#26356e'}
          emissiveIntensity={selected ? 0.22 : 0.08}
          roughness={0.28}
          metalness={0.08}
        />
      </mesh>
      <mesh position={[0, 0, 0.065]}>
        <boxGeometry args={[0.025, 0.29, 0.015]} />
        <meshStandardMaterial color="#eef3ff" roughness={0.4} />
      </mesh>
    </group>
  );
}

function Balcony({ x, y, colors }: { x: number; y: number; colors: SceneColors }) {
  return (
    <group position={[x, y, 0.94]}>
      <mesh position={[0, -0.2, 0.16]} castShadow>
        <boxGeometry args={[0.78, 0.09, 0.54]} />
        <meshStandardMaterial color={colors.floor} roughness={0.58} />
      </mesh>
      <mesh position={[0, 0.02, 0.42]}>
        <boxGeometry args={[0.78, 0.43, 0.035]} />
        <meshStandardMaterial color={colors.body} roughness={0.42} metalness={0.08} />
      </mesh>
      <mesh position={[-0.37, 0.02, 0.16]}>
        <boxGeometry args={[0.035, 0.43, 0.54]} />
        <meshStandardMaterial color={colors.body} roughness={0.42} />
      </mesh>
      <mesh position={[0.37, 0.02, 0.16]}>
        <boxGeometry args={[0.035, 0.43, 0.54]} />
        <meshStandardMaterial color={colors.body} roughness={0.42} />
      </mesh>
    </group>
  );
}

function Scene() {
  const colors = useSceneColors();
  const [selected, setSelected] = useState(0);
  const [hovered, setHovered] = useState<number | null>(null);
  const group = useRef<Group>(null);

  useFrame((_: any, delta: number) => {
    if (group.current) group.current.rotation.y += delta * 0.1;
  });

  const floorY = (floor: number) => 0.42 + floor * 0.73;

  return (
    <>
      <CameraParallax />
      <ambientLight intensity={0.68} color={colors.fillLight} />
      <directionalLight position={[4, 7, 5]} intensity={2.2} color={colors.keyLight} castShadow shadow-mapSize={[1024, 1024]} />
      <directionalLight position={[-4, 3, 2]} intensity={0.75} color={colors.fillLight} />
      <group ref={group}>
        <RoundedBox args={[2.7, 3.02, 1.55]} radius={0.1} smoothness={5} position={[0, 1.5, 0]} castShadow receiveShadow>
          <meshStandardMaterial color={colors.body} roughness={0.63} metalness={0.02} />
        </RoundedBox>
        {[0, 1, 2, 3].map((floor) => (
          <mesh key={`slab-${floor}`} position={[0, 0.38 + floor * 0.73, 0.02]} castShadow>
            <boxGeometry args={[2.82, 0.075, 1.68]} />
            <meshStandardMaterial color={colors.floor} roughness={0.5} />
          </mesh>
        ))}
        {[0, 1, 2, 3].flatMap((floor) => [0, 1].map((side) => {
          const index = floor * 2 + side;
          return (
            <WindowUnit
              key={`front-${index}`}
              position={[side ? 0.68 : -0.68, floorY(floor), 0.81]}
              selected={selected === index}
              hovered={hovered === index}
              colors={colors}
              onSelect={() => setSelected(index)}
              onHover={(value) => setHovered(value ? index : null)}
            />
          );
        }))}
        {[1, 2, 3].map((floor) => (
          <Balcony key={`balcony-${floor}`} x={floor % 2 ? -0.68 : 0.68} y={floorY(floor)} colors={colors} />
        ))}
        <group position={[0, 3.1, 0]}>
          <RoundedBox args={[2.32, 0.22, 1.28]} radius={0.05} smoothness={3} castShadow>
            <meshStandardMaterial color={colors.floor} roughness={0.55} />
          </RoundedBox>
          <mesh position={[0, 0.22, 0]} castShadow>
            <boxGeometry args={[0.92, 0.3, 0.72]} />
            <meshStandardMaterial color={colors.panel} roughness={0.42} />
          </mesh>
        </group>
        <mesh position={[0, 0.22, 0.82]}>
          <boxGeometry args={[0.56, 0.72, 0.06]} />
          <meshStandardMaterial color={colors.panel} roughness={0.34} metalness={0.08} />
        </mesh>
        <mesh position={[0, 0.21, 0.86]}>
          <boxGeometry args={[0.08, 0.76, 0.02]} />
          <meshStandardMaterial color="#eef3ff" roughness={0.38} />
        </mesh>
      </group>
      <mesh rotation={[-Math.PI / 2, 0, 0]} position={[0, -0.04, 0]} receiveShadow>
        <planeGeometry args={[5.8, 4.2]} />
        <meshBasicMaterial color="#dce3ff" transparent opacity={0.28} />
      </mesh>
    </>
  );
}

export function BuildingCanvas() {
  return (
    <div className="lp-building-canvas" aria-label="Интерактивное жилое здание">
      <Canvas
        dpr={[1, 1.5]}
        frameloop="always"
        shadows
        camera={{ position: [5.3, 3.9, 6.9], fov: 31 }}
        gl={{ alpha: true, antialias: true, powerPreference: 'high-performance' }}
      >
        <Scene />
      </Canvas>
    </div>
  );
}
