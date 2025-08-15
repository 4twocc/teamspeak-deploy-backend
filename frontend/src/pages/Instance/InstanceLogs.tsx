import React, { useEffect, useState, useRef } from 'react';
import Button from '@/components/ui/button';
import { RotateCw } from 'lucide-react';
import { toast } from 'sonner';
import { getInstanceLogs } from '@/api/instance';

interface InstanceLogsProps {
  instanceId: string;
}

const InstanceLogs: React.FC<InstanceLogsProps> = ({ instanceId }) => {
  const [loading, setLoading] = useState(false);
  const [logs, setLogs] = useState<string>('');
  const logsEndRef = useRef<HTMLDivElement>(null);

  const fetchLogs = async () => {
    try {
      setLoading(true);
      const response = await getInstanceLogs(instanceId);
      setLogs(response.data);
    } catch (error) {
      console.error(error);
      toast.error('获取实例日志失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchLogs();
  }, [instanceId]);

  useEffect(() => {
    // 滚动到底部
    logsEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [logs]);

  return (
    <div className="p-4">
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-xl font-bold">实例日志</h2>
        <Button 
          onClick={fetchLogs} 
          disabled={loading}
          className="flex items-center gap-2"
        >
          <RotateCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
          刷新
        </Button>
      </div>
      <div className="bg-gray-900 p-4 rounded-lg">
        {logs ? (
          <pre className="text-gray-300 text-sm overflow-auto max-h-96">
            {logs}
          </pre>
        ) : (
          <p className="text-gray-500">
            {loading ? '加载中...' : `暂无实例 ${instanceId} 的日志内容`}
          </p>
        )}
        <div ref={logsEndRef} />
      </div>
    </div>
  );
};

export default InstanceLogs;
