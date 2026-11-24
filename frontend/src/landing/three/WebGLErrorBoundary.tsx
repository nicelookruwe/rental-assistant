import { Component, type ErrorInfo, type ReactNode } from 'react';

export class WebGLErrorBoundary extends Component<{ children: ReactNode; fallback: ReactNode }, { failed: boolean }> {
  state = { failed: false };
  static getDerivedStateFromError() { return { failed: true }; }
  componentDidCatch(error: Error, info: ErrorInfo) { console.error('Landing WebGL scene failed', error, info); }
  render() { return this.state.failed ? this.props.fallback : this.props.children; }
}
