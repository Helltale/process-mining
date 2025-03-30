import React, { useState, useEffect, useRef } from 'react';
import { Button } from '@/components/ui/button';
import DatasetCard from '@/components/DatasetCard';
import { uploadDataset, fetchDatasets, deleteDataset } from '@/services/api';
import { Dataset } from '@/types';

const Home: React.FC = () => {
  const [datasets, setDatasets] = useState<Dataset[]>([]);
  const inputRef = useRef<HTMLInputElement | null>(null);

  useEffect(() => {
    const loadDatasets = async () => {
      try {
        const result = await fetchDatasets();
        setDatasets(result);
      } catch (err) {
        console.error('Ошибка загрузки списка датасетов:', err);
      }
    };

    loadDatasets();
  }, []);

  const simulateProgress = (
    id: string,
    targetProgress: number = 100,
    speed: number = 20
  ) => {
    let current = 0;
    const interval = setInterval(() => {
      current += 5;
      setDatasets((prev) =>
        prev.map((d) =>
          d.id === id
            ? { ...d, progress: Math.min(current, targetProgress) }
            : d
        )
      );

      if (current >= targetProgress) clearInterval(interval);
    }, speed);
  };

  const handleUpload = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;

    const tempId = Date.now().toString();

    const tempDataset: Dataset = {
      id: tempId,
      name: file.name,
      createdAt: new Date().toISOString(),
      uploadedAt: new Date().toISOString(),
      status: 'validating',
      progress: 0,
    };

    setDatasets((prev) => [tempDataset, ...prev]);
    simulateProgress(tempId, 90); // покажем прогресс до 90%, пока сервер отвечает

    try {
      const validatedDataset = await uploadDataset(file, tempDataset);

      setDatasets((prev) =>
        prev.map((d) =>
          d.id === tempId
            ? { ...validatedDataset, progress: 100 }
            : d
        )
      );
    } catch (error) {
      console.error('Ошибка загрузки:', error);
      setDatasets((prev) =>
        prev.map((d) =>
          d.id === tempId ? { ...d, status: 'error', progress: 0 } : d
        )
      );
    }
  };

  const handleButtonClick = () => {
    inputRef.current?.click();
  };

  const handleDelete = async (id: string) => {
    await deleteDataset(id);
    setDatasets((prev) => prev.filter((d) => d.id !== id));
  };

  return (
    <div className="p-6 max-w-screen-lg mx-auto space-y-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold text-white">📁 Мои датасеты</h1>

        <div>
          <input
            ref={inputRef}
            type="file"
            accept=".csv"
            onChange={handleUpload}
            className="hidden"
          />
          <Button
            className="bg-blue-600 text-white rounded-xl hover:bg-blue-700"
            onClick={handleButtonClick}
          >
            Загрузить CSV
          </Button>
        </div>
      </div>

      <div className="grid gap-6 sm:grid-cols-2 md:grid-cols-3">
        {datasets.map((dataset) => (
          <DatasetCard
            key={dataset.id}
            dataset={dataset}
            onDelete={() => handleDelete(dataset.id)}
          />
        ))}
      </div>
    </div>
  );
};

export default Home;
