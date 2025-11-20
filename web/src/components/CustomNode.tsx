import { memo } from 'react';
import { Handle, Position, type NodeProps } from 'reactflow';
import clsx from 'clsx';

const CustomNode = ({ data, selected }: NodeProps) => {
    const { type, id, model, tool_name } = data;

    const getColors = () => {
        switch (type) {
            case 'generate': return 'bg-blue-100 border-blue-500 text-blue-900';
            case 'tool': return 'bg-green-100 border-green-500 text-green-900';
            case 'parallel': return 'bg-purple-100 border-purple-500 text-purple-900';
            case 'selector': return 'bg-orange-100 border-orange-500 text-orange-900';
            case 'group': return 'bg-gray-100 border-gray-500 text-gray-900';
            case 'noop': return 'bg-slate-100 border-slate-500 text-slate-900';
            default: return 'bg-white border-gray-400 text-gray-900';
        }
    };

    return (
        <div className={clsx(
            'px-4 py-2 shadow-md rounded-md border-2 min-w-[150px]',
            getColors(),
            selected ? 'ring-2 ring-offset-1 ring-blue-400' : ''
        )}>
            <Handle type="target" position={Position.Top} className="w-3 h-3 !bg-gray-400" />

            <div className="font-bold text-xs uppercase mb-1 opacity-70">{type}</div>
            <div className="font-semibold text-sm">{id}</div>

            {model && <div className="text-xs mt-1 opacity-80">Model: {model}</div>}
            {tool_name && <div className="text-xs mt-1 opacity-80">Tool: {tool_name}</div>}

            <Handle type="source" position={Position.Bottom} className="w-3 h-3 !bg-gray-400" />
        </div>
    );
};

export default memo(CustomNode);
