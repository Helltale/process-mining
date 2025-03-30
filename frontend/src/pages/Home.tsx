import React, { useState, useEffect, useRef } from 'react';
import { Button } from '@/components/ui/button';
import DatasetCard from '@/components/DatasetCard';
import { uploadDataset, fetchDatasets, deleteDataset, pollValidationProgress } from '@/services/api';
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

  const handleUpload = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;

    const tempId = Date.now().toString();

    const tempDataset: Dataset = {
      id: tempId,
      name: file.name,
      createdAt: new Date().toISOString(),
      uploadedAt: new Date().toISOString(),
      size: file.size,
      status: 'validating',
      progress: 0,
    };

    setDatasets((prev) => [tempDataset, ...prev]);

    try {
      const validatedDataset = await uploadDataset(file, tempDataset);

      // Пуллинг прогресса валидации с сервера
      let progress = 0;
      while (progress < 100) {
        await new Promise((res) => setTimeout(res, 500));
        progress = await pollValidationProgress(validatedDataset.id);

        setDatasets((prev) =>
          prev.map((d) =>
            d.id === tempId ? { ...d, progress, status: 'validating' } : d
          )
        );
      }

      // По завершении обновляем датасет
      setDatasets((prev) =>
        prev.map((d) =>
          d.id === tempId ? { ...validatedDataset, progress: 100, status: 'ready' } : d
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
    try {
      await deleteDataset(id);
      setDatasets((prev) => prev.filter((d) => d.id !== id));
    } catch (error) {
      console.error('Ошибка удаления датасета:', error);
    }
  };

  return (
    <div className="p-6 max-w-screen-lg mx-auto space-y-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold text-white">📁 Датасеты</h1>

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
        {Array.isArray(datasets) && datasets.length > 0 ? (
          datasets.map((dataset) => (
            <DatasetCard
              key={dataset.id}
              dataset={dataset}
              onDelete={() => handleDelete(dataset.id)}
            />
          ))
        ) : (
          <div className="text-white opacity-60 text-sm col-span-full">
            Датасеты не найдены. Загрузите CSV-файл, чтобы начать.
          </div>
        )}
      </div>
    </div>
  );
};

export default Home;
