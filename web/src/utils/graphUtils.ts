import type { DslNode, Workflow } from '../types/dsl';
import { type Node, type Edge } from 'reactflow';
import dagre from 'dagre';

const NODE_WIDTH = 200;
const NODE_HEIGHT = 80;

export const getLayoutedElements = (nodes: Node[], edges: Edge[]) => {
    const dagreGraph = new dagre.graphlib.Graph();
    dagreGraph.setDefaultEdgeLabel(() => ({}));

    dagreGraph.setGraph({ rankdir: 'TB' });

    nodes.forEach((node) => {
        dagreGraph.setNode(node.id, { width: NODE_WIDTH, height: NODE_HEIGHT });
    });

    edges.forEach((edge) => {
        dagreGraph.setEdge(edge.source, edge.target);
    });

    dagre.layout(dagreGraph);

    nodes.forEach((node) => {
        const nodeWithPosition = dagreGraph.node(node.id);
        node.position = {
            x: nodeWithPosition.x - NODE_WIDTH / 2,
            y: nodeWithPosition.y - NODE_HEIGHT / 2,
        };
    });

    return { nodes, edges };
};

export const dslToGraph = (workflow: Workflow) => {
    const nodes: Node[] = [];
    const edges: Edge[] = [];

    const processNode = (node: DslNode) => {
        // Check if node already exists to avoid duplicates if visited multiple times (shouldn't happen with tree traversal but graph might have cycles?)
        // Actually, we iterate the structure.

        const reactFlowNode: Node = {
            id: node.id,
            type: 'custom', // We'll use a generic custom node that renders based on data.type
            position: node.metadata?.position || { x: 0, y: 0 },
            data: { ...node },
        };
        nodes.push(reactFlowNode);

        if (node.next) {
            edges.push({
                id: `${node.id}-${node.next}`,
                source: node.id,
                target: node.next,
                type: 'smoothstep',
            });
        }

        if (node.type === 'selector') {
            node.options.forEach((option, index) => {
                if (option.next) {
                    edges.push({
                        id: `${node.id}-option-${index}`,
                        source: node.id,
                        target: option.next,
                        label: option.case,
                        type: 'smoothstep',
                    });
                }
            });
        }

        if (node.type === 'parallel') {
            node.branches.forEach((branch, branchIndex) => {
                if (branch.nodes.length > 0) {
                    // Connect parallel node to first node of branch
                    edges.push({
                        id: `${node.id}-branch-${branchIndex}`,
                        source: node.id,
                        target: branch.nodes[0].id,
                        type: 'smoothstep',
                        label: `Branch ${branchIndex + 1}`,
                    });

                    // Process branch nodes
                    branch.nodes.forEach(subNode => processNode(subNode));

                    // Connect last node of branch to parallel's next
                    const lastNode = branch.nodes[branch.nodes.length - 1];
                    if (node.next) {
                        // Check if edge already exists (if multiple branches point to same next)
                        // React Flow allows multiple edges.
                        edges.push({
                            id: `${lastNode.id}-${node.next}-merge`,
                            source: lastNode.id,
                            target: node.next,
                            type: 'smoothstep',
                            animated: true, // Visualize merge
                        });
                    }
                }
            });
        }

        if (node.type === 'group') {
            if (node.nodes.length > 0) {
                edges.push({
                    id: `${node.id}-group-start`,
                    source: node.id,
                    target: node.nodes[0].id,
                    type: 'smoothstep',
                });

                node.nodes.forEach(subNode => processNode(subNode));

                const lastNode = node.nodes[node.nodes.length - 1];
                if (node.next) {
                    edges.push({
                        id: `${lastNode.id}-${node.next}-group-end`,
                        source: lastNode.id,
                        target: node.next,
                        type: 'smoothstep',
                    });
                }
            }
        }
    };

    // We need to process top-level nodes
    // But wait, processNode is recursive for parallel/group.
    // What about top-level nodes?
    // The workflow has a list of nodes.
    // If I iterate them, I might process nodes that are already processed (if they were inside a group/parallel)?
    // No, the workflow.nodes list contains the top-level nodes.
    // Parallel/Group nodes contain their children in their own structure, NOT in the top-level list (usually).
    // Let's verify with the YAML structure.
    // The YAML example shows:
    // nodes:
    //   - id: ...
    //   - id: parallel_research
    //     branches:
    //       - nodes: ...
    // So yes, children are nested in the object, not in the main list.

    workflow.nodes.forEach(node => processNode(node));

    // Remove duplicate nodes if any (just in case)
    const uniqueNodes = Array.from(new Map(nodes.map(n => [n.id, n])).values());

    return { nodes: uniqueNodes, edges };
};
