export interface Dataset {
    id: string
    name: string
    uploadedAt: string
    progress: number
    status: 'validating' | 'ready' | 'error'
    error?: string
  }
  