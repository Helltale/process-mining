export interface Dataset {
    id: string;
    name: string;
    uploadedAt: string;
    createdAt: string;
    status: 'pending' | 'validating' | 'ready' | 'error';
    progress: number;
    error?: string;
  }
  