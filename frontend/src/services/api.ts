import { Dataset } from '@/types'

// Используем переменную окружения с fallback
const BASE_URL = import.meta.env.VITE_BACKEND_URL || 'http://localhost:8085';

export const uploadDataset = async (
  file: File,
  tempDataset: Dataset
): Promise<Dataset> => {
  const formData = new FormData();
  formData.append('file', file);

  const res = await fetch(`${BASE_URL}/upload`, {
    method: 'POST',
    body: formData,
  });

  if (!res.ok) {
    const errText = await res.text();
    throw new Error(`Ошибка загрузки: ${res.status} ${errText}`);
  }

  const data = await res.json();

  return {
    id: data.id,
    name: data.name,
    createdAt: data.createdAt,
    uploadedAt: data.uploadedAt,
    status: data.status,
    progress: data.progress,
    size: data.size,
  };
};


export const fetchDatasets = async (): Promise<Dataset[]> => {
  const res = await fetch(`${BASE_URL}/api/datasets`);

  if (!res.ok) {
    const errText = await res.text();
    throw new Error(`Ошибка получения датасетов: ${res.status} ${errText}`);
  }

  const contentType = res.headers.get('content-type') || '';
  if (!contentType.includes('application/json')) {
    const body = await res.text();
    throw new Error(`Ожидался JSON, но получено: ${body}`);
  }

  return res.json();
};

export const buildGraph = async (file: string) => {
    const res = await fetch(`${BASE_URL}/build?file=${encodeURIComponent(file)}`, {
      method: 'POST',
    });
  
    if (!res.ok) {
      throw new Error('Ошибка построения графа');
    }
  };
  

export const deleteDataset = async (id: string) => {
  const res = await fetch(`${BASE_URL}/api/tmp/${id}`, {
    method: 'DELETE',
  });

  if (!res.ok) {
    const errText = await res.text();
    throw new Error(`Ошибка удаления: ${res.status} ${errText}`);
  }
};

export const pollValidationProgress = async (file: string): Promise<number> => {
  const res = await fetch(`${BASE_URL}/api/progress?file=${file}`);
  if (!res.ok) return 0;

  const data = await res.json();
  return data.progress ?? 0;
};
