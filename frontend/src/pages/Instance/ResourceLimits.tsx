// src/pages/Instance/ResourceLimits.tsx
import React, { useEffect, useState } from 'react';
import { Card } from '@/components/ui/card';
import Button from '@/components/ui/button';
import { RotateCw } from 'lucide-react';
import { toast } from 'sonner';
import { getInstance } from '@/api/instance';
import type { TSInstance } from '@/types/Instance';

const ResourceLimits: React.FC<{ instanceId: string }> = ({ instanceId }) => {
  const [loading, setLoading] = useState(false);
  const [instance, setInstance] = useState<TSInstance | null>(null);

  const fetchInstance = async () => {
    try {
      setLoading(true);
      const response = await getInstance(instanceId);
      setInstance(response.data);
    } catch (error) {
      console.error(error);
      toast.error('获取实例资源限制失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchInstance();
  }, [instanceId]);

  return (
    <Card className="p-6">
      <div className="flex justify-between items-center mb-4">
        <h3 className="text-lg font-semibold">资源限制</h3>
        <Button 
          onClick={fetchInstance} 
          disabled={loading}
          className="flex items-center gap-2"
        >
          <RotateCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
          刷新
        </Button>
      </div>
      
      <div className="space-y-4">
        <div>
          <div className="flex justify-between mb-1">
            <span className="text-sm font-medium">最大客户端数</span>
            <span className="text-sm text-gray-500">{instance?.config.max_clients || 'N/A'}</span>
          </div>
        </div>
        
        <div>
          <div className="flex justify-between mb-1">
            <span className="text-sm font-medium">语音端口</span>
            <span className="text-sm text-gray-500">{instance?.config.voice_port || 'N/A'}</span>
          </div>
        </div>
        
        <div>
          <div className="flex justify-between mb-1">
            <span className="text-sm font-medium">文件传输端口</span>
            <span className="text-sm text-gray-500">{instance?.config.file_port || 'N/A'}</span>
          </div>
        </div>
        
        <div>
          <div className="flex justify-between mb-1">
            <span className="text-sm font-medium">查询端口</span>
            <span className="text-sm text-gray-500">{instance?.config.query_port || 'N/A'}</span>
          </div>
        </div>
      </div>
    </Card>
  );
};

export default ResourceLimits;
