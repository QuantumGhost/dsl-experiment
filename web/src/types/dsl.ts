export type NodeType = 'generate' | 'tool' | 'parallel' | 'selector' | 'group' | 'noop';

export interface BaseNode {
    id: string;
    type: NodeType;
    next?: string;
    metadata?: {
        position?: { x: number; y: number };
        [key: string]: any;
    };
    [key: string]: any;
}

export interface GenerateNode extends BaseNode {
    type: 'generate';
    model: string;
    system_prompt?: string;
    temperature?: number;
    context_mode?: string;
}

export interface ToolNode extends BaseNode {
    type: 'tool';
    tool_name: string;
    parameters?: Record<string, string>;
}

export interface ParallelNode extends BaseNode {
    type: 'parallel';
    branches: { nodes: DslNode[] }[];
}

export interface SelectorNode extends BaseNode {
    type: 'selector';
    model: string;
    system_prompt?: string;
    options: { case: string; next: string }[];
}

export interface GroupNode extends BaseNode {
    type: 'group';
    nodes: DslNode[];
}

export interface NoopNode extends BaseNode {
    type: 'noop';
}

export type DslNode = GenerateNode | ToolNode | ParallelNode | SelectorNode | GroupNode | NoopNode;

export interface Workflow {
    name: string;
    description?: string;
    nodes: DslNode[];
}
