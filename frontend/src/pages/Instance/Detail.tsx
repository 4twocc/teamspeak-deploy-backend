// src/pages/Instance/Detail.tsx
import React from 'react';
import InstanceInfo from './InstanceInfo';
import InstanceLogs from './InstanceLogs';
import InstanceResources from './InstanceResources';
import ResourceLimits from './ResourceLimits';
import { Card } from '@/components/ui/card';

const InstanceDetail: React.FC<{ instanceId: string }> = ({ instanceId }) => {
  const [activeTab, setActiveTab] = React.useState('info');

  return (
    <Card className="p-6">
      <div className="border-b border-gray-200">
        <nav className="-mb-px flex space-x-8">
          <button
            className={`py-2 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'info'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
            onClick={() => setActiveTab('info')}
          >
            实例信息
          </button>
          <button
            className={`py-2 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'resources'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
            onClick={() => setActiveTab('resources')}
          >
            资源监控
          </button>
          <button
            className={`py-2 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'logs'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
            onClick={() => setActiveTab('logs')}
          >
            日志
          </button>
        </nav>
      </div>
      <div className="py-4">
        {activeTab === 'info' && <InstanceInfo instanceId={instanceId} />}
        {activeTab === 'resources' && (
          <div>
            <InstanceResources instanceId={instanceId} />
            <div className="mt-4">
              <ResourceLimits instanceId={instanceId} />
            </div>
          </div>
        )}
        {activeTab === 'logs' && <InstanceLogs instanceId={instanceId} />}
      </div>
    </Card>
  );
};

export default InstanceDetail;
