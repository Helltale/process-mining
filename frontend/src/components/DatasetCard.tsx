import React from 'react';
import { Button } from '@/components/ui/button';
import { Progress } from '@/components/ui/progress';
import { Dataset } from '@/types';
import { Trash, Play } from 'lucide-react';
import { useNavigate } from 'react-router-dom';

interface Props {
  dataset: Dataset;
  onDelete: () => void;
}

const DatasetCard: React.FC<Props> = ({ dataset, onDelete }) => {
  const navigate = useNavigate();

  return (
    <div className="rounded-2xl bg-[#161b22] text-white p-4 shadow-md space-y-4 border border-[#30363d] transition hover:scale-[1.02] duration-200">
      <div className="flex justify-between items-start">
        <div>
          <div className="font-semibold text-lg">{dataset.name}</div>
          <div className="text-sm text-gray-400">
            Загружено: {new Date(dataset.uploadedAt).toLocaleString()}
          </div>
        </div>
        <Button
          variant="outline"
          size="icon"
          onClick={onDelete}
          className="border border-red-500 hover:bg-red-500/20 text-red-500 rounded-full"
        >
          <Trash className="w-4 h-4" />
        </Button>
      </div>

      <Progress
        value={dataset.progress || 0}
        className="h-2 rounded-full bg-gray-700"
      />

      {dataset.status === 'validating' && (
        <div className="text-sm text-blue-400">Валидация...</div>
      )}

      {dataset.status === 'error' && (
        <div className="text-sm text-red-400">
          Ошибка: {dataset.error || 'Неизвестная ошибка'}
        </div>
      )}

      {dataset.status === 'ready' && (
        <Button
          onClick={() => navigate(`/graph?file=${dataset.name}`)}
          className="flex items-center gap-2 bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-xl"
        >
          <Play className="w-4 h-4" />
          Построить граф
        </Button>
      )}
    </div>
  );
};

export default DatasetCard;
