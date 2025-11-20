import { useCallback } from 'react';
import ReactFlow, {
  Background,
  Controls,
  MiniMap,
  ReactFlowProvider,
  type NodeTypes,
  type OnSelectionChangeParams,
} from 'reactflow';
import 'reactflow/dist/style.css';
import { useAppStore } from './store/useAppStore';
import CustomNode from './components/CustomNode';
import Sidebar from './components/Sidebar';

const nodeTypes: NodeTypes = {
  custom: CustomNode,
};

const Flow = () => {
  const {
    nodes,
    edges,
    onNodesChange,
    onEdgesChange,
    onConnect,
    selectNode,
    layoutGraph,
  } = useAppStore();

  const onSelectionChange = useCallback(({ nodes }: OnSelectionChangeParams) => {
    if (nodes.length > 0) {
      selectNode(nodes[0].id);
    } else {
      selectNode(null);
    }
  }, [selectNode]);

  return (
    <div className="flex-1 h-full relative">
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onConnect={onConnect}
        onSelectionChange={onSelectionChange}
        nodeTypes={nodeTypes}
        fitView
      >
        <Background />
        <Controls />
        <MiniMap />
      </ReactFlow>
      <div className="absolute top-4 left-4 z-10">
        <button
          onClick={layoutGraph}
          className="bg-white p-2 rounded shadow text-sm font-medium hover:bg-gray-50 border border-gray-200"
        >
          Auto Layout
        </button>
      </div>
    </div>
  );
};

function App() {
  return (
    <div className="flex h-screen w-screen overflow-hidden bg-gray-50">
      <ReactFlowProvider>
        <Flow />
        <Sidebar />
      </ReactFlowProvider>
    </div>
  );
}

export default App;
