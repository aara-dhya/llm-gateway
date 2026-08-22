<script lang="ts">
    import type { RequestLog } from './+page.server';

    interface Props {
        data: {
            metrics: RequestLog[];
        };
    }

    let { data }: Props = $props();

    // Svelte 5 $derived runes: Dynamically recompute analytics whenever metrics update
    let totalRequests = $derived(data.metrics.length);
    let totalTokens = $derived(data.metrics.reduce((acc, m) => acc + m.total_tokens, 0));
    let promptTokens = $derived(data.metrics.reduce((acc, m) => acc + m.prompt_tokens, 0));
    let completionTokens = $derived(data.metrics.reduce((acc, m) => acc + m.completion_tokens, 0));

    // Estimated cost benchmark ($0.0015 / 1k prompt tokens, $0.0020 / 1k completion tokens)
    let estimatedCost = $derived(
        ((promptTokens / 1000) * 0.0015 + (completionTokens / 1000) * 0.002).toFixed(4)
    );
</script>

<div class="min-h-screen bg-gray-900 text-gray-100 p-8 font-sans">
    <div class="max-w-7xl mx-auto space-y-8">
        
        <!-- Header Bar -->
        <div class="flex items-center justify-between border-b border-gray-800 pb-5">
            <div>
                <h1 class="text-3xl font-bold tracking-tight text-white">LLM Observability Gateway</h1>
                <p class="text-sm text-gray-400 mt-1">Real-time token metrics, request logs, and cost analytics</p>
            </div>
            <div class="flex items-center space-x-2">
                <span class="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                    <span class="w-2 h-2 mr-2 bg-emerald-400 rounded-full animate-pulse"></span> Gateway Active
                </span>
            </div>
        </div>

        <!-- Analytics Overview Cards -->
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            <!-- Total Requests Card -->
            <div class="bg-gray-800/60 border border-gray-700/60 rounded-xl p-6 shadow-sm">
                <p class="text-xs font-semibold uppercase tracking-wider text-gray-400">Total Requests</p>
                <p class="text-3xl font-extrabold text-white mt-2">{totalRequests}</p>
                <p class="text-xs text-gray-500 mt-1">Logged API calls</p>
            </div>

            <!-- Total Tokens Card -->
            <div class="bg-gray-800/60 border border-gray-700/60 rounded-xl p-6 shadow-sm">
                <p class="text-xs font-semibold uppercase tracking-wider text-gray-400">Total Tokens</p>
                <p class="text-3xl font-extrabold text-indigo-400 mt-2">{totalTokens.toLocaleString()}</p>
                <p class="text-xs text-gray-500 mt-1">Prompt + Completion</p>
            </div>

            <!-- Token Breakdown Card -->
            <div class="bg-gray-800/60 border border-gray-700/60 rounded-xl p-6 shadow-sm">
                <p class="text-xs font-semibold uppercase tracking-wider text-gray-400">Prompt / Completion</p>
                <p class="text-xl font-bold text-white mt-2">
                    <span class="text-sky-400">{promptTokens.toLocaleString()}</span> / <span class="text-purple-400">{completionTokens.toLocaleString()}</span>
                </p>
                <p class="text-xs text-gray-500 mt-1">Input vs. Output tokens</p>
            </div>

            <!-- Estimated Cost Card -->
            <div class="bg-gray-800/60 border border-gray-700/60 rounded-xl p-6 shadow-sm">
                <p class="text-xs font-semibold uppercase tracking-wider text-gray-400">Est. Total Cost</p>
                <p class="text-3xl font-extrabold text-emerald-400 mt-2">${estimatedCost}</p>
                <p class="text-xs text-gray-500 mt-1">Calculated API expenditure</p>
            </div>
        </div>

        <!-- Requests Log Table -->
        <div class="bg-gray-800/60 border border-gray-700/60 rounded-xl shadow-sm overflow-hidden">
            <div class="px-6 py-4 border-b border-gray-700/60 flex justify-between items-center">
                <h2 class="text-lg font-semibold text-white">Recent Requests</h2>
                <span class="text-xs text-gray-400">Showing last {data.metrics.length} records</span>
            </div>

            <div class="overflow-x-auto">
                <table class="w-full text-left text-sm text-gray-300">
                    <thead class="bg-gray-900/50 text-xs uppercase font-medium text-gray-400 border-b border-gray-700/60">
                        <tr>
                            <th class="px-6 py-3">ID</th>
                            <th class="px-6 py-3">Timestamp</th>
                            <th class="px-6 py-3">Model</th>
                            <th class="px-6 py-3">Prompt Tokens</th>
                            <th class="px-6 py-3">Completion Tokens</th>
                            <th class="px-6 py-3">Total Tokens</th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-gray-700/40">
                        {#if data.metrics.length === 0}
                            <tr>
                                <td colspan="6" class="px-6 py-8 text-center text-gray-500">
                                    No requests recorded yet. Send a request through the gateway to view live metrics!
                                </td>
                            </tr>
                        {:else}
                            {#each data.metrics as log (log.id)}
                                <tr class="hover:bg-gray-700/30 transition-colors">
                                    <td class="px-6 py-4 font-mono text-xs text-gray-400">#{log.id}</td>
                                    <td class="px-6 py-4 text-xs whitespace-nowrap text-gray-400">
                                        {new Date(log.timestamp).toLocaleString()}
                                    </td>
                                    <td class="px-6 py-4 font-medium text-white">
                                        <span class="inline-flex items-center px-2 py-0.5 rounded text-xs bg-gray-700 text-gray-200">
                                            {log.model || 'unknown'}
                                        </span>
                                    </td>
                                    <td class="px-6 py-4 text-sky-400 font-mono">{log.prompt_tokens}</td>
                                    <td class="px-6 py-4 text-purple-400 font-mono">{log.completion_tokens}</td>
                                    <td class="px-6 py-4 font-bold text-indigo-300 font-mono">{log.total_tokens}</td>
                                </tr>
                            {/each}
                        {/if}
                    </tbody>
                </table>
            </div>
        </div>

    </div>
</div>