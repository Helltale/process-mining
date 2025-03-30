// src/types/cytoscape-svg.d.ts

declare module 'cytoscape-svg' {
    import type { Core } from 'cytoscape';
  
    interface SvgOptions {
      scale?: number;
      full?: boolean;
      bg?: string;
      xmlns?: string;
    }
  
    export default function cytoscapeSvg(cytoscape: any): void;
  
    declare module 'cytoscape' {
      interface Core {
        svg(options?: SvgOptions): string;
      }
    }
  }
  