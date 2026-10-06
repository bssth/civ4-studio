import {editor} from "../wailsjs/go/models";

// Plot types and flags, see editor/wbmap_plots.go and editor/wbmap_view.go
export const PLOT_PEAK = 0;
export const PLOT_HILLS = 1;
export const PLOT_LAND = 2;
export const PLOT_OCEAN = 3;

export const FLAG_N_OF_RIVER = 1;
export const FLAG_W_OF_RIVER = 2;
export const FLAG_START = 4;
export const FLAG_IMPROVEMENT = 8;
export const FLAG_ROUTE = 16;
export const FLAG_LANDMARK = 32;

const terrainColors: Record<string, string> = {
    TERRAIN_GRASS: '#4e8b2e',
    TERRAIN_PLAINS: '#a49a45',
    TERRAIN_DESERT: '#e3d39a',
    TERRAIN_TUNDRA: '#8c8a76',
    TERRAIN_SNOW: '#f0f3f5',
    TERRAIN_MARSH: '#5b7a58',
    TERRAIN_COAST: '#3b7fc4',
    TERRAIN_OCEAN: '#1d4f91',
    TERRAIN_LAKE: '#3b7fc4',
    TERRAIN_PEAK: '#7a7a7a',
    TERRAIN_HILL: '#8a7a55',
};

const featureColors: Record<string, string> = {
    FEATURE_FOREST: '#1f5f1f',
    FEATURE_JUNGLE: '#0d4a33',
    FEATURE_ICE: '#ffffff',
    FEATURE_OASIS: '#40c0a0',
    FEATURE_FLOOD_PLAINS: '#c0b060',
    FEATURE_FALLOUT: '#a0ff40',
};

// Stable color for types that are not in the tables above (e.g. added by mods)
function hashColor(type: string, saturation = 35, lightness = 50): string {
    let h = 0;
    for (let i = 0; i < type.length; i++) h = (h * 31 + type.charCodeAt(i)) % 360;
    return `hsl(${h}, ${saturation}%, ${lightness}%)`;
}

export function terrainColor(type: string | undefined): string {
    if (!type) return '#000000';
    return terrainColors[type] ?? hashColor(type);
}

export function featureColor(type: string): string {
    return featureColors[type] ?? hashColor(type, 50, 30);
}

export interface Layers {
    rivers: boolean;
    resources: boolean;
    cities: boolean;
    units: boolean;
    starts: boolean;
    grid: boolean;
}

export interface StartMarker {
    player: number;
    x: number;
    y: number;
    color: string;
    // The player starts at a random location, the position is not used by the game
    random: boolean;
}

/** Edge of a plot where a river can be: south is isNOfRiver, east is isWOfRiver */
export interface RiverEdge {
    x: number;
    y: number;
    side: 'south' | 'east';
}

export interface RenderOptions {
    cell: number;
    layers: Layers;
    ownerColor: (owner: number) => string;
    starts: StartMarker[];
    selected: { x: number, y: number } | null;
    // Square of plots the brush covers under the cursor
    brush?: { x: number, y: number, size: number } | null;
    // River edge under the cursor in river mode
    edge?: RiverEdge | null;
    // Map column shown at the left edge of a map wrapping east-west (0 shows the map as it is stored)
    offset?: number;
    // The map wraps east-west: the brush continues on the other side of the seam
    wrapX?: boolean;
}

/** Remainder that is never negative, e.g. mod(-1, 10) = 9 */
export function mod(a: number, n: number): number {
    return ((a % n) + n) % n;
}

/** Screen column of a map column when the map is shifted by offset columns */
export function screenColumn(x: number, width: number, offset: number): number {
    return mod(x - offset, width);
}

/** Map column shown in a screen column, inverse of screenColumn */
export function mapColumn(column: number, width: number, offset: number): number {
    return mod(column + offset, width);
}

/**
 * Plots covered by a square brush of the given size centered at x, y, clipped to the map.
 * On a map wrapping east-west the brush continues on the other side of the seam.
 */
export function brushCells(view: editor.MapView, x: number, y: number, size: number, wrapX = false): { x: number, y: number }[] {
    const r = Math.floor((size - 1) / 2);
    const cells = [];
    const seen = new Set<number>();
    for (let dy = -r; dy <= r; dy++) {
        for (let dx = -r; dx <= r; dx++) {
            const cx = wrapX ? mod(x + dx, view.width) : x + dx, cy = y + dy;
            if (cx >= 0 && cy >= 0 && cx < view.width && cy < view.height && !seen.has(cy * view.width + cx)) {
                // A brush wider than a narrow map must not cover a plot twice
                seen.add(cy * view.width + cx);
                cells.push({x: cx, y: cy});
            }
        }
    }
    return cells;
}

/** Plots on the line between two plots, so fast mouse moves do not leave gaps */
export function lineCells(x0: number, y0: number, x1: number, y1: number): { x: number, y: number }[] {
    const cells = [];
    const dx = Math.abs(x1 - x0), dy = -Math.abs(y1 - y0);
    const sx = x0 < x1 ? 1 : -1, sy = y0 < y1 ? 1 : -1;
    let err = dx + dy, x = x0, y = y0;
    for (; ;) {
        cells.push({x, y});
        if (x === x1 && y === y1) break;
        const e2 = 2 * err;
        if (e2 >= dy) {
            err += dy;
            x += sx;
        }
        if (e2 <= dx) {
            err += dx;
            y += sy;
        }
    }
    return cells;
}

/**
 * River edge nearest to a point inside a plot. Edges of the neighbour plots are stored there:
 * the north edge is the south edge of the plot above, the west edge is the east edge of the plot on the left.
 */
export function nearestEdge(view: editor.MapView, x: number, y: number, fx: number, fy: number, wrapX = false): RiverEdge | null {
    // fx, fy are 0..1 inside the plot, fy grows downwards on the screen
    const distances: [number, RiverEdge][] = [
        [1 - fy, {x, y, side: 'south'}],
        [fy, {x, y: y + 1, side: 'south'}],
        [1 - fx, {x, y, side: 'east'}],
        // On a wrapping map the west edge of the first column is the east edge of the last one
        [fx, {x: wrapX ? mod(x - 1, view.width) : x - 1, y, side: 'east'}],
    ];
    distances.sort((a, b) => a[0] - b[0]);
    for (const [, edge] of distances) {
        if (edge.x >= 0 && edge.y >= 0 && edge.x < view.width && edge.y < view.height) return edge;
    }
    return null;
}

/** Canvas row of a map row: the game counts rows from the bottom */
export function rowOf(view: editor.MapView, y: number): number {
    return view.height - 1 - y;
}

export function drawMap(ctx: CanvasRenderingContext2D, view: editor.MapView, o: RenderOptions) {
    const {cell, layers} = o;
    const w = view.width, h = view.height;
    const offset = o.offset ?? 0;
    // Left pixel of a map column
    const left = (x: number) => screenColumn(x, w, offset) * cell;
    ctx.clearRect(0, 0, w * cell, h * cell);

    const terrainFill = view.terrains.map(terrainColor);
    const featureFill = view.features.map(featureColor);

    for (let y = 0; y < h; y++) {
        const py = rowOf(view, y) * cell;
        for (let x = 0; x < w; x++) {
            const i = y * w + x;
            const px = left(x);
            const t = view.terrain[i];
            ctx.fillStyle = t >= 0 ? terrainFill[t] : '#000000';
            ctx.fillRect(px, py, cell, cell);

            const plotType = view.plot_type[i];
            if (plotType === PLOT_PEAK) {
                ctx.fillStyle = '#6e6e6e';
                ctx.beginPath();
                ctx.moveTo(px, py + cell);
                ctx.lineTo(px + cell / 2, py);
                ctx.lineTo(px + cell, py + cell);
                ctx.fill();
                if (cell >= 6) {
                    ctx.fillStyle = '#ffffff';
                    ctx.beginPath();
                    ctx.moveTo(px + cell * 0.35, py + cell * 0.3);
                    ctx.lineTo(px + cell / 2, py);
                    ctx.lineTo(px + cell * 0.65, py + cell * 0.3);
                    ctx.fill();
                }
            } else if (plotType === PLOT_HILLS) {
                ctx.fillStyle = 'rgba(60, 40, 10, 0.35)';
                ctx.beginPath();
                ctx.ellipse(px + cell / 2, py + cell, cell / 2, cell * 0.55, 0, Math.PI, 2 * Math.PI);
                ctx.fill();
            }

            const f = view.feature[i];
            if (f >= 0) {
                ctx.fillStyle = featureFill[f];
                if (view.features[f] === 'FEATURE_ICE') {
                    ctx.globalAlpha = 0.8;
                    ctx.fillRect(px, py, cell, cell);
                    ctx.globalAlpha = 1;
                } else {
                    // Dotted overlay, so the terrain stays visible
                    const d = Math.max(1, Math.floor(cell / 4));
                    for (let dy = d / 2; dy < cell; dy += d * 2) {
                        for (let dx = (dy / d) % 2 ? d * 1.5 : d / 2; dx < cell; dx += d * 2) {
                            ctx.fillRect(px + dx, py + dy, d, d);
                        }
                    }
                }
            }
        }
    }

    if (layers.grid && cell >= 6) {
        ctx.strokeStyle = 'rgba(0, 0, 0, 0.15)';
        ctx.lineWidth = 1;
        ctx.beginPath();
        for (let x = 0; x <= w; x++) {
            ctx.moveTo(x * cell + 0.5, 0);
            ctx.lineTo(x * cell + 0.5, h * cell);
        }
        for (let y = 0; y <= h; y++) {
            ctx.moveTo(0, y * cell + 0.5);
            ctx.lineTo(w * cell, y * cell + 0.5);
        }
        ctx.stroke();
    }

    if (layers.rivers) {
        ctx.strokeStyle = '#4fb3ff';
        ctx.lineWidth = Math.max(1, cell / 6);
        ctx.beginPath();
        for (let i = 0; i < view.flags.length; i++) {
            const flags = view.flags[i];
            if (!(flags & (FLAG_N_OF_RIVER | FLAG_W_OF_RIVER))) continue;
            const x = i % w, y = Math.floor(i / w);
            const px = left(x), py = rowOf(view, y) * cell;
            if (flags & FLAG_N_OF_RIVER) {
                // The river runs along the southern edge of the plot
                ctx.moveTo(px, py + cell);
                ctx.lineTo(px + cell, py + cell);
            }
            if (flags & FLAG_W_OF_RIVER) {
                // The river runs along the eastern edge of the plot
                ctx.moveTo(px + cell, py);
                ctx.lineTo(px + cell, py + cell);
            }
        }
        ctx.stroke();
    }

    if (offset !== 0) {
        // The seam: the first column of the map as it is stored
        ctx.strokeStyle = 'rgba(255, 255, 255, 0.6)';
        ctx.lineWidth = 1;
        ctx.setLineDash([6, 4]);
        ctx.beginPath();
        ctx.moveTo(left(0) + 0.5, 0);
        ctx.lineTo(left(0) + 0.5, h * cell);
        ctx.stroke();
        ctx.setLineDash([]);
    }

    for (let i = 0; i < view.flags.length; i++) {
        const x = i % w, y = Math.floor(i / w);
        const px = left(x), py = rowOf(view, y) * cell;
        const flags = view.flags[i];

        if (layers.resources && view.bonus[i] >= 0 && cell >= 4) {
            const r = Math.max(1.5, cell / 6);
            ctx.fillStyle = '#ffd54f';
            ctx.strokeStyle = '#5d4000';
            ctx.lineWidth = 1;
            ctx.beginPath();
            ctx.arc(px + cell * 0.75, py + cell * 0.25, r, 0, 2 * Math.PI);
            ctx.fill();
            if (cell >= 8) ctx.stroke();
        }

        if (layers.units && view.unit_count[i] > 0) {
            const s = Math.max(2, cell * 0.4);
            ctx.fillStyle = o.ownerColor(view.unit_owner[i]);
            ctx.strokeStyle = '#000000';
            ctx.lineWidth = 1;
            ctx.beginPath();
            ctx.moveTo(px + 1, py + cell - 1);
            ctx.lineTo(px + 1 + s, py + cell - 1);
            ctx.lineTo(px + 1 + s / 2, py + cell - 1 - s);
            ctx.closePath();
            ctx.fill();
            if (cell >= 8) ctx.stroke();
        }

        if (layers.cities && view.city_owner[i] >= 0) {
            const m = Math.max(1, cell * 0.15);
            ctx.fillStyle = o.ownerColor(view.city_owner[i]);
            ctx.fillRect(px + m, py + m, cell - 2 * m, cell - 2 * m);
            ctx.strokeStyle = '#ffffff';
            ctx.lineWidth = Math.max(1, cell / 8);
            ctx.strokeRect(px + m, py + m, cell - 2 * m, cell - 2 * m);
        }

        if (layers.starts && (flags & FLAG_START) && cell >= 4) {
            // A start for a random civilization
            ctx.strokeStyle = '#ffffff';
            ctx.lineWidth = Math.max(1, cell / 8);
            ctx.beginPath();
            ctx.arc(px + cell / 2, py + cell / 2, cell * 0.3, 0, 2 * Math.PI);
            ctx.stroke();
        }
    }

    if (layers.starts) {
        for (const s of o.starts) {
            const cx = left(s.x) + cell / 2, cy = rowOf(view, s.y) * cell + cell / 2;
            const r = Math.max(4, cell * 0.6);
            ctx.globalAlpha = s.random ? 0.45 : 1;
            ctx.fillStyle = s.color;
            ctx.strokeStyle = '#000000';
            ctx.lineWidth = 2;
            ctx.setLineDash(s.random ? [3, 2] : []);
            ctx.beginPath();
            ctx.arc(cx, cy, r, 0, 2 * Math.PI);
            ctx.fill();
            ctx.stroke();
            ctx.setLineDash([]);
            if (r >= 7) {
                ctx.fillStyle = '#000000';
                ctx.font = `bold ${Math.round(r)}px sans-serif`;
                ctx.textAlign = 'center';
                ctx.textBaseline = 'middle';
                ctx.fillText(String(s.player), cx, cy + 1);
            }
            ctx.globalAlpha = 1;
        }
    }

    if (o.brush) {
        const cells = brushCells(view, o.brush.x, o.brush.y, o.brush.size, o.wrapX);
        if (cells.length > 0) {
            const top = Math.max(...cells.map(c => c.y)), bottom = Math.min(...cells.map(c => c.y));
            // A brush crossing the edge of the screen is drawn as two parts
            const columns = [...new Set(cells.map(c => screenColumn(c.x, w, offset)))].sort((a, b) => a - b);
            ctx.strokeStyle = '#ffffff';
            ctx.lineWidth = 2;
            ctx.setLineDash([4, 3]);
            let start = 0;
            for (let i = 1; i <= columns.length; i++) {
                if (i === columns.length || columns[i] !== columns[i - 1] + 1) {
                    ctx.strokeRect(columns[start] * cell, rowOf(view, top) * cell,
                        (columns[i - 1] - columns[start] + 1) * cell, (top - bottom + 1) * cell);
                    start = i;
                }
            }
            ctx.setLineDash([]);
        }
    }

    if (o.edge) {
        const px = left(o.edge.x), py = rowOf(view, o.edge.y) * cell;
        ctx.strokeStyle = '#ffeb3b';
        ctx.lineWidth = Math.max(3, cell / 4);
        ctx.beginPath();
        if (o.edge.side === 'south') {
            ctx.moveTo(px, py + cell);
            ctx.lineTo(px + cell, py + cell);
        } else {
            ctx.moveTo(px + cell, py);
            ctx.lineTo(px + cell, py + cell);
        }
        ctx.stroke();
    }

    if (o.selected) {
        const px = left(o.selected.x), py = rowOf(view, o.selected.y) * cell;
        ctx.strokeStyle = '#ffeb3b';
        ctx.lineWidth = Math.max(2, cell / 6);
        ctx.strokeRect(px, py, cell, cell);
    }
}
