export interface Dataset {
    id: string;
    name: string;
    uploadedAt: string;
    createdAt: string;
    size: number;
    status: 'pending' | 'validating' | 'ready' | 'error';
    progress: number;
    error?: string;
  }
  