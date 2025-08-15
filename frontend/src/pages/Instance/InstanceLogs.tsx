import React from 'react';

interface InstanceLogsProps {
  instanceId: string;
}

const InstanceLogs: React.FC<InstanceLogsProps> = ({ instanceId }) => {
  return (
    <div className="p-4">
      <h2 className="text-xl font-bold mb-4">实例日志</h2>
      <div className="bg-gray-100 p-4 rounded-lg">
        <p className="text-gray-500">实例 {instanceId} 的日志内容将在这里显示</p>
      </div>
    </div>
  );
};

export default InstanceLogs;