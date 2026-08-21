import type { ServerLoadEvent } from '@sveltejs/kit';

export interface RequestLog {
    id: number;
    timestamp: string;
    model: string;
    prompt_tokens: number;
    completion_tokens: number;
    total_tokens: number;
}

// We explicitly type the 'event' parameter using SvelteKit's core types
export const load = async (event: ServerLoadEvent) => {
    const { fetch } = event; // Extract fetch from the event

    try {
        const res = await fetch('http://localhost:8080/api/metrics');
        
        if (!res.ok) {
            console.error("Backend returned an error:", res.status);
            return { metrics: [] as RequestLog[] };
        }

        const metrics: RequestLog[] = await res.json();
        return { metrics };
    } catch (error) {
        console.error("Failed to connect to Go backend:", error);
        return { metrics: [] as RequestLog[] }; 
    }
};