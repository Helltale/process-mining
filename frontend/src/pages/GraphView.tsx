import React, { useEffect, useState, useRef } from 'react';
import { useSearchParams } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { Slider } from '@/components/ui/slider';
import { Download, RotateCw } from 'lucide-react';
import cytoscape from 'cytoscape';
import svg from 'cytoscape-svg'

svg(cytoscape) // <- регистрируем плагин


const GraphView: React.FC = () => {
  const [searchParams] = useSearchParams();
  const [graphData, setGraphData] = useState<any>(null);
  const [threshold, setThreshold] = useState(100);
  const containerRef = useRef<HTMLDivElement>(null);
  const cyRef = useRef<cytoscape.Core | null>(null);

  const file = searchParams.get('file');

  useEffect(() => {
    const fetchGraph = async () => {
      const res = await fetch(`/graph?file=${file}`);
      const data = await res.json();
      setGraphData(data);
    };

    if (file) fetchGraph();
  }, [file]);

  useEffect(() => {
    if (!graphData || !containerRef.current) return;

    const counts = graphData.edges.map((e: any) => e.data.count);
    const min = Math.min(...counts);
    const max = Math.max(...counts);
    const edgeThreshold = min + ((max - min) * (100 - threshold)) / 100;

    const elements = [
      ...graphData.nodes.map((n: any) => ({ data: n.data })),
      ...graphData.edges
        .filter((e: any) => e.data.count >= edgeThreshold)
        .map((e: any) => ({
          data: e.data,
          classes: e.data.style === 'dashed' ? 'dashed' : '',
        })),
    ];

    if (cyRef.current) cyRef.current.destroy();

    const cy = cytoscape({
      container: containerRef.current,
      elements,
      style: [
        {
          selector: 'node',
          style: {
            label: 'data(label)',
            'background-color': 'data(color)',
            'text-valign': 'center',
            'text-halign': 'center',
            'font-size': 10,
          },
        },
        {
          selector: 'edge',
          style: {
            width: 2,
            label: 'data(label)',
            'curve-style': 'bezier',
            'target-arrow-shape': 'triangle',
            'arrow-scale': 0.8,
            'font-size': 8,
            'line-color': '#999',
            'target-arrow-color': '#999',
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
    const svg = cyRef.current.svg({ full: true });
    const blob = new Blob([svg], { type: 'image/svg+xml;charset=utf-8' });
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
        <h2 className="text-xl font-bold">Граф процесса для: {file}</h2>
        <div className="flex gap-2">
          <Button onClick={downloadSVG} variant="outline">
            <Download className="w-4 h-4" /> Скачать SVG
          </Button>
          <Button onClick={() => setThreshold(100)} variant="ghost">
            <RotateCw className="w-4 h-4" /> Сброс
          </Button>
        </div>
      </div>

      <div className="flex items-center gap-4">
        <label className="font-medium">Порог мощности:</label>
        <Slider value={[threshold]} onValueChange={(v) => setThreshold(v[0])} min={0} max={100} className="w-64" />
        <span>{threshold}%</span>
      </div>

      <div
        ref={containerRef}
        className="border rounded bg-white min-h-[600px] h-[70vh] w-full"
      />
    </div>
  );
};

export default GraphView;
