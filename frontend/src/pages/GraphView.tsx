import React, { useEffect, useRef, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { Slider } from '@/components/ui/slider';
import { Download, RotateCw } from 'lucide-react';
import cytoscape from 'cytoscape';
import svg from 'cytoscape-svg';

svg(cytoscape);

const GraphView: React.FC = () => {
  const [searchParams] = useSearchParams();
  const [graphData, setGraphData] = useState<any>(null);
  const [threshold, setThreshold] = useState(100);
  const [isLoading, setIsLoading] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);
  const cyRef = useRef<cytoscape.Core | null>(null);

  const file = searchParams.get('file');

  const clean = (s: string) => s?.replace(/\r/g, '').trim();

  useEffect(() => {
    const fetchGraph = async () => {
      setIsLoading(true);
      try {
        const res = await fetch(`http://localhost:8085/graph?file=${file}`);
        const contentType = res.headers.get('content-type');
        if (!contentType?.includes('application/json')) {
          const text = await res.text();
          throw new Error(`Сервер вернул не JSON: ${text}`);
        }

        const raw = await res.json();

        const nodes: any[] = [];
        const nodeIds = new Set<string>();

        raw.nodes.forEach((n: any) => {
          const id = clean(n.data.id);
          nodeIds.add(id);
          nodes.push({
            data: {
              ...n.data,
              id,
              label: `${clean(n.data.label)}\n(${n.data.count})`,
              color: n.data.color || 'gray',
            },
          });
        });

        const edges: any[] = raw.edges
          .map((e: any) => {
            const source = clean(e.data.source);
            const target = clean(e.data.target);
            if (!nodeIds.has(source) || !nodeIds.has(target)) return null;
            return {
              data: {
                id: `${source}_${target}`,
                source,
                target,
                label: clean(e.data.label),
                count: e.data.count,
              },
              classes:
                e.data.style === 'dashed' ||
                source === 'start' ||
                target === 'end'
                  ? 'dashed'
                  : '',
            };
          })
          .filter(Boolean);

        setGraphData({ nodes, edges });
      } catch (err) {
        console.error('Ошибка загрузки графа:', err);
      } finally {
        setIsLoading(false);
      }
    };

    if (file) fetchGraph();
  }, [file]);

  useEffect(() => {
    if (!graphData || !containerRef.current) return;

    const counts = graphData.edges.map((e: any) => e.data.count || 0);
    const min = Math.min(...counts);
    const max = Math.max(...counts);
    const edgeThreshold = min + ((max - min) * (100 - threshold)) / 100;

    const elements = [
      ...graphData.nodes,
      ...graphData.edges.filter((e: any) => e.data.count >= edgeThreshold),
    ];

    if (cyRef.current) cyRef.current.destroy();

    const cy = cytoscape({
      container: containerRef.current,
      elements,
      style: [
        {
          selector: 'node[color]',
          style: {
            'background-color': 'data(color)',
            'shape': 'round-rectangle',
            'label': 'data(label)',
            'text-wrap': 'wrap',
            'text-max-width': '100px',
            'text-valign': 'center',
            'text-halign': 'center',
            'font-size': 10,
            'padding': '8px',
            'min-width': '50px',
            'min-height': '30px',
          },
        },
        {
          selector: 'edge',
          style: {
            width: 2,
            label: 'data(label)',
            'curve-style': 'bezier',
            'target-arrow-shape': 'triangle',
            'target-arrow-color': '#999',
            'line-color': '#999',
            'arrow-scale': 0.8,
            'font-size': 8,
          },
        },
        {
          selector: '.dashed',
          style: {
            'line-style': 'dashed',
          },
        },
      ],
      layout: {
        name: 'breadthfirst',
        directed: true,
        padding: 20,
      },
    });

    cyRef.current = cy;
  }, [graphData, threshold]);

  const downloadSVG = () => {
    if (!cyRef.current) return;
    const svgStr = cyRef.current.svg({ full: true });
    const blob = new Blob([svgStr], { type: 'image/svg+xml;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${file || 'graph'}.svg`;
    a.click();
    URL.revokeObjectURL(url);
  };

  return (
    <div className="p-6 space-y-4 max-w-screen-xl mx-auto">
      <div className="flex justify-between items-center">
        <h2 className="text-xl font-bold text-white">Граф процесса для: {file}</h2>
        <div className="flex gap-2">
          <Button onClick={downloadSVG} variant="outline">
            <Download className="w-4 h-4 mr-1" /> Скачать SVG
          </Button>
          <Button onClick={() => setThreshold(100)} variant="ghost">
            <RotateCw className="w-4 h-4 mr-1" /> Сброс
          </Button>
        </div>
      </div>

      <div className="flex items-center gap-4">
        <label className="font-medium text-white">Порог мощности:</label>
        <Slider
          value={[threshold]}
          onValueChange={(v) => setThreshold(v[0])}
          min={0}
          max={100}
          className="w-64"
        />
        <span className="text-white">{threshold}%</span>
      </div>

      <div
        ref={containerRef}
        className="border rounded bg-white min-h-[600px] h-[70vh] w-full relative"
      >
        {isLoading && (
          <div className="absolute inset-0 flex items-center justify-center bg-white/80 z-10">
            <div className="animate-spin rounded-full h-16 w-16 border-4 border-blue-500 border-t-transparent"></div>
          </div>
        )}
      </div>
    </div>
  );
};

export default GraphView;
