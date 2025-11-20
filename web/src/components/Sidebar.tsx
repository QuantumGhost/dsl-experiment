import React from 'react';
import { useAppStore } from '../store/useAppStore';
import type { DslNode, NodeType, GenerateNode, ToolNode, ParallelNode, SelectorNode, GroupNode, NoopNode } from '../types/dsl';

const createDefaultNode = (type: NodeType): DslNode => {
    const id = `${type}_${Date.now()}`;
    const base = { id, type };
    switch (type) {
        case 'generate': return { ...base, model: 'gpt-4o', system_prompt: 'You are a helpful assistant.', temperature: 0.7 } as GenerateNode;
        case 'tool': return { ...base, tool_name: 'weather', parameters: {} } as ToolNode;
        case 'parallel': return { ...base, branches: [{ nodes: [] }, { nodes: [] }] } as ParallelNode;
        case 'selector': return { ...base, model: 'gpt-4o', options: [{ case: 'default', next: '' }] } as SelectorNode;
        case 'group': return { ...base, nodes: [] } as GroupNode;
        case 'noop': return { ...base } as NoopNode;
    }
    return base as DslNode;
};

const Sidebar = () => {
    const { selectedNodeId, nodes, updateNodeData, workflow, getYaml, loadYaml, addNode, deleteNode } = useAppStore();

    const selectedNode = nodes.find(n => n.id === selectedNodeId)?.data as DslNode | undefined;

    const handleChange = (field: string, value: any) => {
        if (selectedNodeId) {
            updateNodeData(selectedNodeId, { [field]: value });
        }
    };

    const handleImport = (e: React.ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0];
        if (file) {
            const reader = new FileReader();
            reader.onload = (e) => {
                const content = e.target?.result as string;
                if (content) loadYaml(content);
            };
            reader.readAsText(file);
        }
    };

    const handleExport = () => {
        const yamlContent = getYaml();
        const blob = new Blob([yamlContent], { type: 'text/yaml' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = 'workflow.yaml';
        a.click();
    };

    const handleAddNode = (type: NodeType) => {
        addNode(createDefaultNode(type));
    };

    const handleDeleteNode = () => {
        if (selectedNodeId) {
            if (confirm('Are you sure you want to delete this node?')) {
                deleteNode(selectedNodeId);
            }
        }
    };

    if (!selectedNode) {
        return (
            <div className="w-80 bg-white border-l border-gray-200 p-4 flex flex-col h-full overflow-y-auto">
                <h2 className="text-lg font-bold mb-4">Workflow</h2>
                <div className="mb-4">
                    <label className="block text-sm font-medium text-gray-700">Name</label>
                    <input
                        type="text"
                        className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm border p-2"
                        value={workflow.name}
                        readOnly
                    />
                </div>

                <div className="mb-6">
                    <h3 className="text-sm font-medium text-gray-700 mb-2">Add Node</h3>
                    <div className="grid grid-cols-2 gap-2">
                        {(['generate', 'tool', 'selector', 'parallel', 'group', 'noop'] as NodeType[]).map(type => (
                            <button
                                key={type}
                                onClick={() => handleAddNode(type)}
                                className="px-3 py-2 bg-gray-100 hover:bg-gray-200 rounded text-sm font-medium capitalize border border-gray-300"
                            >
                                {type}
                            </button>
                        ))}
                    </div>
                </div>

                <div className="mt-auto space-y-2">
                    <label className="block w-full">
                        <span className="sr-only">Import YAML</span>
                        <input type="file" accept=".yaml,.yml" onChange={handleImport} className="block w-full text-sm text-slate-500
                  file:mr-4 file:py-2 file:px-4
                  file:rounded-full file:border-0
                  file:text-sm file:font-semibold
                  file:bg-violet-50 file:text-violet-700
                  hover:file:bg-violet-100
                "/>
                    </label>
                    <button onClick={handleExport} className="w-full py-2 px-4 bg-blue-600 text-white rounded hover:bg-blue-700">
                        Export YAML
                    </button>
                </div>
            </div>
        );
    }

    return (
        <div className="w-80 bg-white border-l border-gray-200 p-4 flex flex-col h-full overflow-y-auto">
            <div className="flex justify-between items-center mb-4">
                <h2 className="text-lg font-bold">Edit Node</h2>
                <button
                    onClick={handleDeleteNode}
                    className="text-red-600 hover:text-red-800 text-sm font-medium"
                >
                    Delete
                </button>
            </div>

            <div className="space-y-4">
                <div>
                    <label className="block text-sm font-medium text-gray-700">ID</label>
                    <input
                        type="text"
                        value={selectedNode.id}
                        disabled
                        className="mt-1 block w-full rounded-md border-gray-300 bg-gray-100 shadow-sm sm:text-sm border p-2"
                    />
                </div>

                <div>
                    <label className="block text-sm font-medium text-gray-700">Type</label>
                    <input
                        type="text"
                        value={selectedNode.type}
                        disabled
                        className="mt-1 block w-full rounded-md border-gray-300 bg-gray-100 shadow-sm sm:text-sm border p-2"
                    />
                </div>

                <div>
                    <label className="block text-sm font-medium text-gray-700">Next Node ID</label>
                    <input
                        type="text"
                        value={selectedNode.next || ''}
                        onChange={(e) => handleChange('next', e.target.value)}
                        className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm border p-2"
                    />
                </div>

                {selectedNode.type === 'generate' && (
                    <>
                        <div>
                            <label className="block text-sm font-medium text-gray-700">Model</label>
                            <input
                                type="text"
                                value={(selectedNode as any).model || ''}
                                onChange={(e) => handleChange('model', e.target.value)}
                                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm border p-2"
                            />
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-gray-700">Temperature</label>
                            <input
                                type="number"
                                step="0.1"
                                value={(selectedNode as any).temperature ?? 0.7}
                                onChange={(e) => handleChange('temperature', parseFloat(e.target.value))}
                                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm border p-2"
                            />
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-gray-700">System Prompt</label>
                            <textarea
                                value={(selectedNode as any).system_prompt || ''}
                                onChange={(e) => handleChange('system_prompt', e.target.value)}
                                rows={4}
                                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm border p-2"
                            />
                        </div>
                    </>
                )}

                {selectedNode.type === 'tool' && (
                    <>
                        <div>
                            <label className="block text-sm font-medium text-gray-700">Tool Name</label>
                            <input
                                type="text"
                                value={(selectedNode as any).tool_name || ''}
                                onChange={(e) => handleChange('tool_name', e.target.value)}
                                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm border p-2"
                            />
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-gray-700">Parameters (JSON)</label>
                            <textarea
                                value={JSON.stringify((selectedNode as any).parameters || {}, null, 2)}
                                onChange={(e) => {
                                    try {
                                        const params = JSON.parse(e.target.value);
                                        handleChange('parameters', params);
                                    } catch (err) {
                                        // ignore invalid json while typing
                                    }
                                }}
                                rows={4}
                                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm border p-2 font-mono"
                            />
                        </div>
                    </>
                )}

                {selectedNode.type === 'selector' && (
                    <>
                        <div>
                            <label className="block text-sm font-medium text-gray-700">Model</label>
                            <input
                                type="text"
                                value={(selectedNode as any).model || ''}
                                onChange={(e) => handleChange('model', e.target.value)}
                                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm border p-2"
                            />
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-gray-700">System Prompt</label>
                            <textarea
                                value={(selectedNode as any).system_prompt || ''}
                                onChange={(e) => handleChange('system_prompt', e.target.value)}
                                rows={4}
                                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm border p-2"
                            />
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-gray-700">Options</label>
                            <div className="space-y-2 mt-1">
                                {((selectedNode as any).options || []).map((opt: any, idx: number) => (
                                    <div key={idx} className="flex gap-2 items-center">
                                        <input
                                            type="text"
                                            placeholder="Case"
                                            value={opt.case}
                                            onChange={(e) => {
                                                const newOptions = [...(selectedNode as any).options];
                                                newOptions[idx] = { ...opt, case: e.target.value };
                                                handleChange('options', newOptions);
                                            }}
                                            className="w-1/3 rounded border p-1 text-sm"
                                        />
                                        <input
                                            type="text"
                                            placeholder="Next ID"
                                            value={opt.next}
                                            onChange={(e) => {
                                                const newOptions = [...(selectedNode as any).options];
                                                newOptions[idx] = { ...opt, next: e.target.value };
                                                handleChange('options', newOptions);
                                            }}
                                            className="w-1/3 rounded border p-1 text-sm"
                                        />
                                        <button
                                            onClick={() => {
                                                const newOptions = (selectedNode as any).options.filter((_: any, i: number) => i !== idx);
                                                handleChange('options', newOptions);
                                            }}
                                            className="text-red-500 hover:text-red-700"
                                        >
                                            ×
                                        </button>
                                    </div>
                                ))}
                                <button
                                    onClick={() => {
                                        const newOptions = [...((selectedNode as any).options || []), { case: '', next: '' }];
                                        handleChange('options', newOptions);
                                    }}
                                    className="text-sm text-blue-600 hover:text-blue-800"
                                >
                                    + Add Option
                                </button>
                            </div>
                        </div>
                    </>
                )}
            </div>
            <div className="pt-4 border-t border-gray-200 mt-4">
                <details>
                    <summary className="text-sm font-medium text-gray-700 cursor-pointer">Advanced: Raw JSON</summary>
                    <textarea
                        value={JSON.stringify(selectedNode, null, 2)}
                        onChange={(e) => {
                            try {
                                const data = JSON.parse(e.target.value);
                                if (data.id !== selectedNode.id) {
                                    return;
                                }
                                updateNodeData(selectedNode.id, data);
                            } catch (err) {
                                // ignore
                            }
                        }}
                        rows={10}
                        className="mt-2 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-xs font-mono border p-2"
                    />
                </details>
            </div>
        </div>
    );
};

export default Sidebar;
