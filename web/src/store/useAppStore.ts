import { create } from 'zustand';
import type { Workflow, DslNode } from '../types/dsl';
import { type Node, type Edge, type NodeChange, type EdgeChange, type Connection, applyNodeChanges, applyEdgeChanges } from 'reactflow';
import { dslToGraph, getLayoutedElements } from '../utils/graphUtils';
import yaml from 'js-yaml';

interface AppState {
	workflow: Workflow;
	nodes: Node[];
	edges: Edge[];
	selectedNodeId: string | null;

	setWorkflow: (workflow: Workflow) => void;
	loadYaml: (yamlString: string) => void;
	getYaml: () => string;

	onNodesChange: (changes: NodeChange[]) => void;
	onEdgesChange: (changes: EdgeChange[]) => void;
	onConnect: (connection: Connection) => void;

	selectNode: (id: string | null) => void;
	updateNodeData: (id: string, data: Partial<DslNode>) => void;

	layoutGraph: () => void;
	addNode: (node: DslNode) => void;
	deleteNode: (id: string) => void;
}

export const useAppStore = create<AppState>((set, get) => ({
	workflow: { name: 'New Workflow', nodes: [] },
	nodes: [],
	edges: [],
	selectedNodeId: null,

	setWorkflow: (workflow) => {
		const { nodes, edges } = dslToGraph(workflow);
		set({ workflow, nodes, edges });
	},

	loadYaml: (yamlString) => {
		try {
			const workflow = yaml.load(yamlString) as Workflow;
			if (!workflow || typeof workflow !== 'object') throw new Error('Invalid YAML');
			if (!workflow.nodes) workflow.nodes = [];

			const { nodes, edges } = dslToGraph(workflow);
			const layouted = getLayoutedElements(nodes, edges);

			set({ workflow, nodes: layouted.nodes, edges: layouted.edges });
		} catch (e) {
			console.error('Failed to parse YAML', e);
			alert('Failed to parse YAML: ' + (e as Error).message);
		}
	},

	getYaml: () => {
		const { workflow, nodes } = get();

		const positionMap = new Map(nodes.map(n => [n.id, n.position]));

		const updateNodeMetadata = (node: DslNode) => {
			const pos = positionMap.get(node.id);
			if (pos) {
				node.metadata = { ...node.metadata, position: pos };
			}

			if (node.type === 'parallel') {
				(node as any).branches.forEach((b: any) => b.nodes.forEach(updateNodeMetadata));
			}
			if (node.type === 'group') {
				(node as any).nodes.forEach(updateNodeMetadata);
			}
		};

		const newWorkflow = JSON.parse(JSON.stringify(workflow)); // Deep copy
		newWorkflow.nodes.forEach(updateNodeMetadata);

		return yaml.dump(newWorkflow);
	},

	onNodesChange: (changes) => {
		set({
			nodes: applyNodeChanges(changes, get().nodes),
		});
	},

	onEdgesChange: (changes) => {
		set({
			edges: applyEdgeChanges(changes, get().edges),
		});
	},

	onConnect: (connection) => {
		const { workflow, nodes } = get();
		if (!connection.source || !connection.target) return;

		// Sync positions first
		const positionMap = new Map(nodes.map(n => [n.id, n.position]));
		const syncPositions = (n: DslNode) => {
			const pos = positionMap.get(n.id);
			if (pos) {
				n.metadata = { ...n.metadata, position: pos };
			}
			if (n.type === 'parallel') {
				(n as any).branches.forEach((b: any) => b.nodes.forEach(syncPositions));
			}
			if (n.type === 'group') {
				(n as any).nodes.forEach(syncPositions);
			}
		};

		const newWorkflow = JSON.parse(JSON.stringify(workflow));
		newWorkflow.nodes.forEach(syncPositions);

		// Update the source node's 'next' field
		const updateNext = (nodeList: DslNode[]): boolean => {
			for (let i = 0; i < nodeList.length; i++) {
				if (nodeList[i].id === connection.source) {
					nodeList[i].next = connection.target || undefined;
					return true;
				}
				if (nodeList[i].type === 'parallel') {
					for (const branch of (nodeList[i] as any).branches) {
						if (updateNext(branch.nodes)) return true;
					}
				}
				if (nodeList[i].type === 'group') {
					if (updateNext((nodeList[i] as any).nodes)) return true;
				}
			}
			return false;
		};

		updateNext(newWorkflow.nodes);

		// Regenerate graph
		const { nodes: newNodes, edges: newEdges } = dslToGraph(newWorkflow);
		const posMap = new Map(nodes.map(n => [n.id, n.position]));
		const mergedNodes = newNodes.map(n => ({
			...n,
			position: posMap.get(n.id) || n.position
		}));

		set({ workflow: newWorkflow, nodes: mergedNodes, edges: newEdges });
	},

	selectNode: (id) => set({ selectedNodeId: id }),

	updateNodeData: (id, data) => {
		const { workflow, nodes } = get();

		const updateDslNode = (nodeList: DslNode[]): boolean => {
			for (let i = 0; i < nodeList.length; i++) {
				if (nodeList[i].id === id) {
					nodeList[i] = { ...nodeList[i], ...data } as DslNode;
					return true;
				}
				if (nodeList[i].type === 'parallel') {
					for (const branch of (nodeList[i] as any).branches) {
						if (updateDslNode(branch.nodes)) return true;
					}
				}
				if (nodeList[i].type === 'group') {
					if (updateDslNode((nodeList[i] as any).nodes)) return true;
				}
			}
			return false;
		};

		const newWorkflow = JSON.parse(JSON.stringify(workflow));
		updateDslNode(newWorkflow.nodes);

		let newEdges = get().edges;
		if (data.next !== undefined || (data as any).options !== undefined || (data as any).branches !== undefined || (data as any).nodes !== undefined) {
			const graph = dslToGraph(newWorkflow);
			newEdges = graph.edges;

			const posMap = new Map(nodes.map(n => [n.id, n.position]));
			const mergedNodes = graph.nodes.map(n => ({
				...n,
				position: posMap.get(n.id) || n.position
			}));

			set({ workflow: newWorkflow, nodes: mergedNodes, edges: newEdges });
		} else {
			const newNodes = nodes.map(n => {
				if (n.id === id) {
					return { ...n, data: { ...n.data, ...data } };
				}
				return n;
			});
			set({ workflow: newWorkflow, nodes: newNodes });
		}
	},

	addNode: (node: DslNode) => {
		const { workflow, nodes } = get();

		// Sync current positions to workflow before modifying
		const positionMap = new Map(nodes.map(n => [n.id, n.position]));
		const syncPositions = (n: DslNode) => {
			const pos = positionMap.get(n.id);
			if (pos) {
				n.metadata = { ...n.metadata, position: pos };
			}
			if (n.type === 'parallel') {
				(n as any).branches.forEach((b: any) => b.nodes.forEach(syncPositions));
			}
			if (n.type === 'group') {
				(n as any).nodes.forEach(syncPositions);
			}
		};

		const newWorkflow = JSON.parse(JSON.stringify(workflow));
		newWorkflow.nodes.forEach(syncPositions);

		// Add new node
		newWorkflow.nodes.push(node);

		const { nodes: newNodes, edges: newEdges } = dslToGraph(newWorkflow);
		// We don't auto-layout here to preserve positions, unless it's a new node which defaults to 0,0
		// But dslToGraph uses metadata.position if available.

		set({ workflow: newWorkflow, nodes: newNodes, edges: newEdges });
	},

	deleteNode: (id: string) => {
		const { workflow, nodes } = get();

		// Sync positions
		const positionMap = new Map(nodes.map(n => [n.id, n.position]));
		const syncPositions = (n: DslNode) => {
			const pos = positionMap.get(n.id);
			if (pos) {
				n.metadata = { ...n.metadata, position: pos };
			}
			if (n.type === 'parallel') {
				(n as any).branches.forEach((b: any) => b.nodes.forEach(syncPositions));
			}
			if (n.type === 'group') {
				(n as any).nodes.forEach(syncPositions);
			}
		};

		const newWorkflow = JSON.parse(JSON.stringify(workflow));
		newWorkflow.nodes.forEach(syncPositions);

		// Recursive delete
		const deleteFromList = (list: DslNode[]): boolean => {
			const index = list.findIndex(n => n.id === id);
			if (index !== -1) {
				list.splice(index, 1);
				return true;
			}
			for (const node of list) {
				if (node.type === 'parallel') {
					for (const branch of (node as any).branches) {
						if (deleteFromList(branch.nodes)) return true;
					}
				}
				if (node.type === 'group') {
					if (deleteFromList((node as any).nodes)) return true;
				}
			}
			return false;
		};

		deleteFromList(newWorkflow.nodes);

		const { nodes: newNodes, edges: newEdges } = dslToGraph(newWorkflow);
		set({ workflow: newWorkflow, nodes: newNodes, edges: newEdges, selectedNodeId: null });
	},

	layoutGraph: () => {
		const { nodes, edges } = get();
		const layouted = getLayoutedElements(nodes, edges);
		set({ nodes: [...layouted.nodes], edges: [...layouted.edges] });
	}
}));
