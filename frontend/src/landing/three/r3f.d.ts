declare module '@react-three/fiber' {
  export const Canvas: any;
  export const useFrame: any;
  export const useThree: any;
}

declare module '@react-three/drei' {
  export const RoundedBox: any;
  export const OrbitControls: any;
}

declare namespace JSX {
  interface IntrinsicElements {
    [element: string]: any;
  }
}
