// src/pages/Instance/InstanceResources.tsx
import React, { useEffect, useState } from 'react';
import Button from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { RotateCw } from 'lucide-react';
import { toast } from 'sonner';
import type { ResourceUsage } from '@/types/Instance';

const InstanceResources: React.FC<{ instanceId: string }> = ({ instanceId }) => {
  const [loading, setLoading] = useState(false);
  const [resources, setResources] = useState<ResourceUsage>();

  const fetchResources = async () => {
    try {
      setLoading(true);
      // 这里应该调用实际的API获取资源数据
      // const data = await getInstanceResources(instanceId);
      // 模拟数据
      setResources({
        cpu_percent: Math.random() * 100,
        memory_mb: Math.random() * 512,
        memory_percent: Math.random() * 100,
        disk_usage_mb: Math.random() * 2048,
        network_in: Math.random() * 102400,
        network_out: Math.random() * 51200,
        timestamp: new Date().toISOString()
      });
    } catch (error) {
      console.error(error)
      toast.error('获取资源使用情况失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchResources();
    const timer = setInterval(fetchResources, 30000); // 每30秒刷新一次
    
    return () => clearInterval(timer);
  }, [instanceId]);

  const getStatus = (percent: number) => {
    if (percent > 80) return 'error';
    if (percent > 60) return 'warning';
    return 'success';
  };

  return (
    <Card className="p-6">
      <div className="flex justify-between items-center mb-4">
        <h3 className="text-lg font-semibold">资源使用情况</h3>
        <Button 
          onClick={fetchResources} 
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
            <span className="text-sm font-medium">CPU</span>
            <span className="text-sm text-gray-500">
              {resources?.cpu_percent !== undefined ? `${resources.cpu_percent.toFixed(1)}%` : 'N/A'}
            </span>
          </div>
          <div className="w-full bg-gray-200 rounded-full h-2.5">
            <div 
              className={`h-2.5 rounded-full ${
                getStatus(resources?.cpu_percent || 0) === 'error' ? 'bg-red-600' : 
                getStatus(resources?.cpu_percent || 0) === 'warning' ? 'bg-yellow-500' : 'bg-green-600'
              }`} 
              style={{ width: `${resources?.cpu_percent}%` }}
            ></div>
          </div>
        </div>
        
        <div>
          <div className="flex justify-between mb-1">
            <span className="text-sm font-medium">内存</span>
            <span className="text-sm text-gray-500">
              {resources?.memory_mb !== undefined ? `${resources.memory_mb.toFixed(1)} MB (${resources?.memory_percent.toFixed(1)}%)` : 'N/A'}
            </span>
          </div>
          <div className="w-full bg-gray-200 rounded-full h-2.5">
            <div 
              className={`h-2.5 rounded-full ${
                getStatus(resources?.memory_percent || 0) === 'error' ? 'bg-red-600' : 
                getStatus(resources?.memory_percent || 0) === 'warning' ? 'bg-yellow-500' : 'bg-green-600'
              }`} 
              style={{ width: `${resources?.memory_percent}%` }}
            ></div>
          </div>
        </div>
        
        <div>
          <div className="flex justify-between mb-1">
            <span className="text-sm font-medium">磁盘</span>
            <span className="text-sm text-gray-500">
              {resources?.disk_usage_mb !== undefined ? `${(resources.disk_usage_mb / 1024).toFixed(2)} GB` : 'N/A'}
            </span>
          </div>
          <div className="w-full bg-gray-200 rounded-full h-2.5">
            <div 
              className="h-2.5 rounded-full bg-blue-600" 
              style={{ width: `${Math.min(100, (resources?.disk_usage_mb || 0) / 5120 * 100)}%` }}
            ></div>
          </div>
        </div>
        
        <div className="grid grid-cols-2 gap-4">
          <div className="bg-gray-50 p-3 rounded-lg">
            <div className="text-sm text-gray-500">网络流入</div>
            <div className="font-medium">
              {resources?.network_in !== undefined ? `${(resources.network_in / 1024).toFixed(2)} KB` : 'N/A'}
            </div>
          </div>
          <div className="bg-gray-50 p-3 rounded-lg">
            <div className="text-sm text-gray-500">网络流出</div>
            <div className="font-medium">
              {resources?.network_out !== undefined ? `${(resources.network_out / 1024).toFixed(2)} KB` : 'N/A'}
            </div>
          </div>
        </div>
      </div>
    </Card>
  );
};

export default InstanceResources;
