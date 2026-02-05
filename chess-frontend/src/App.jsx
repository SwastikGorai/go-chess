import React, { useState, useEffect, useMemo } from 'react';
import {
  Settings,
  RefreshCw,
  Moon,
  Sun,
  Trophy,
  Palette,
  LayoutGrid,
  Ghost,
  X,
  Flag,
  AlertTriangle,
  Globe,
  Copy,
  Users,
  Eye,
  Play,
  ArrowRight,
} from 'lucide-react';
import './App.css';

// --- CHESS LOGIC & HELPERS ---

const INITIAL_BOARD = [
  ['r', 'n', 'b', 'q', 'k', 'b', 'n', 'r'],
  ['p', 'p', 'p', 'p', 'p', 'p', 'p', 'p'],
  [null, null, null, null, null, null, null, null],
  [null, null, null, null, null, null, null, null],
  [null, null, null, null, null, null, null, null],
  [null, null, null, null, null, null, null, null],
  ['P', 'P', 'P', 'P', 'P', 'P', 'P', 'P'],
  ['R', 'N', 'B', 'Q', 'K', 'B', 'N', 'R'],
];

const isWhite = (piece) => piece === piece.toUpperCase();
const isBlack = (piece) => piece === piece.toLowerCase();
const getPieceColor = (piece) => (piece ? (isWhite(piece) ? 'white' : 'black') : null);

const getValidMoves = (board, row, col, checkSafety = true) => {
  const piece = board[row][col];
  if (!piece) return [];

  const moves = [];
  const color = getPieceColor(piece);

  const directions = {
    pawn: [],
    rook: [
      [0, 1],
      [0, -1],
      [1, 0],
      [-1, 0],
    ],
    bishop: [
      [1, 1],
      [1, -1],
      [-1, 1],
      [-1, -1],
    ],
    knight: [
      [2, 1],
      [2, -1],
      [-2, 1],
      [-2, -1],
      [1, 2],
      [1, -2],
      [-1, 2],
      [-1, -2],
    ],
    queen: [
      [0, 1],
      [0, -1],
      [1, 0],
      [-1, 0],
      [1, 1],
      [1, -1],
      [-1, 1],
      [-1, -1],
    ],
    king: [
      [0, 1],
      [0, -1],
      [1, 0],
      [-1, 0],
      [1, 1],
      [1, -1],
      [-1, 1],
      [-1, -1],
    ],
  };

  const type = piece.toLowerCase();

  if (type === 'p') {
    const dir = color === 'white' ? -1 : 1;
    const startRow = color === 'white' ? 6 : 1;

    // Move forward 1
    if (!board[row + dir]?.[col]) {
      moves.push({ r: row + dir, c: col });
      // Move forward 2
      if (row === startRow && !board[row + dir * 2]?.[col]) {
        moves.push({ r: row + dir * 2, c: col });
      }
    }
    // Captures
    [
      [dir, 1],
      [dir, -1],
    ].forEach(([dr, dc]) => {
      const r = row + dr,
        c = col + dc;
      if (board[r]?.[c] && getPieceColor(board[r][c]) !== color) {
        moves.push({ r, c });
      }
    });
  } else if (type === 'n' || type === 'k') {
    (type === 'n' ? directions.knight : directions.king).forEach(([dr, dc]) => {
      const r = row + dr,
        c = col + dc;
      if (r >= 0 && r < 8 && c >= 0 && c < 8) {
        const target = board[r][c];
        if (!target || getPieceColor(target) !== color) {
          moves.push({ r, c });
        }
      }
    });
  } else {
    // Sliding pieces (R, B, Q)
    const dirs = type === 'r' ? directions.rook : type === 'b' ? directions.bishop : directions.queen;
    dirs.forEach(([dr, dc]) => {
      let r = row + dr,
        c = col + dc;
      while (r >= 0 && r < 8 && c >= 0 && c < 8) {
        const target = board[r][c];
        if (!target) {
          moves.push({ r, c });
        } else {
          if (getPieceColor(target) !== color) moves.push({ r, c });
          break;
        }
        r += dr;
        c += dc;
      }
    });
  }

  if (checkSafety) {
    return moves.filter((move) => {
      const newBoard = board.map((r) => [...r]);
      newBoard[move.r][move.c] = piece;
      newBoard[row][col] = null;
      return !isKingInCheck(newBoard, color);
    });
  }

  return moves;
};

const findKing = (board, color) => {
  const king = color === 'white' ? 'K' : 'k';
  for (let r = 0; r < 8; r++) {
    for (let c = 0; c < 8; c++) {
      if (board[r][c] === king) return { r, c };
    }
  }
  return null;
};

const isKingInCheck = (board, color) => {
  const kingPos = findKing(board, color);
  if (!kingPos) return false;

  const opponentColor = color === 'white' ? 'black' : 'white';

  for (let r = 0; r < 8; r++) {
    for (let c = 0; c < 8; c++) {
      const piece = board[r][c];
      if (piece && getPieceColor(piece) === opponentColor) {
        const moves = getValidMoves(board, r, c, false);
        if (moves.some((m) => m.r === kingPos.r && m.c === kingPos.c)) {
          return true;
        }
      }
    }
  }
  return false;
};

// --- BACKEND API HELPERS ---

const API_BASE_URL =
  import.meta.env.VITE_API_BASE_URL ||
  (typeof window !== 'undefined' && window.CHESS_API_BASE_URL) ||
  'http://localhost:8080';

const fetchJSON = async (path, options = {}) => {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    headers: { 'Content-Type': 'application/json', ...(options.headers || {}) },
    ...options,
  });

  let payload = null;
  try {
    payload = await response.json();
  } catch (e) {
    payload = null;
  }

  if (!response.ok) {
    const message = payload?.error || `Request failed with ${response.status}`;
    const error = new Error(message);
    error.status = response.status;
    throw error;
  }

  return payload;
};

const FILES = 'abcdefgh';

const squareFromCoords = (r, c) => `${FILES[c]}${8 - r}`;

const coordsFromSquare = (square) => {
  const file = square[0].toLowerCase();
  const rank = Number.parseInt(square[1], 10);
  return { r: 8 - rank, c: FILES.indexOf(file) };
};

const parseFEN = (fen) => {
  const [boardPart, activeColor] = fen.trim().split(/\s+/);
  const rows = boardPart.split('/');
  const board = rows.map((row) => {
    const squares = [];
    for (const char of row) {
      if (char >= '1' && char <= '8') {
        const count = Number.parseInt(char, 10);
        for (let i = 0; i < count; i++) squares.push(null);
      } else {
        squares.push(char);
      }
    }
    return squares;
  });
  return {
    board,
    turn: activeColor === 'b' ? 'black' : 'white',
  };
};

// --- ASSETS & THEMES ---

const BOARD_THEMES = {
  wood: { light: '#eedca5', dark: '#b58863', name: 'Classic Wood', border: 'border-[#8b6547]' },
  emerald: { light: '#f0f4e6', dark: '#769656', name: 'Tournament Green', border: 'border-[#587240]' },
  ocean: { light: '#e0ecf7', dark: '#86a6bf', name: 'Ocean Mist', border: 'border-[#5e7c91]' },
  midnight: { light: '#aeb1b5', dark: '#585b5e', name: 'Midnight Slate', border: 'border-[#36383a]' },
  grayscale: { light: '#e0e0e0', dark: '#888888', name: 'Newspaper', border: 'border-[#404040]' },
};

const Pieces = {
  Classic: {
    P: (props) => (
      <svg viewBox="0 0 45 45" {...props}>
        <path
          d="M22.5 9c-2.21 0-4 1.79-4 4 0 .89.29 1.71.78 2.38C17.33 16.5 16 18.59 16 21c0 2.03.94 3.84 2.41 5.03-3 1.06-7.41 5.55-7.41 13.47h23c0-7.92-4.41-12.41-7.41-13.47 1.47-1.19 2.41-3 2.41-5.03 0-2.41-1.33-4.5-3.28-5.62.49-.67.78-1.49.78-2.38 0-2.21-1.79-4-4-4z"
          strokeWidth="1.5"
          strokeLinecap="round"
        />
      </svg>
    ),
    R: (props) => (
      <svg viewBox="0 0 45 45" {...props}>
        <path
          d="M9 39h27v-3H9v3zM12 36v-4h21v4H12zM11 14V9h4v2h5V9h5v2h5V9h4v5h-2.932L22.5 19.5 13.932 14H11zM12 32h21V19.5L22.5 30 12 19.5V32z"
          strokeWidth="1.5"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>
    ),
    N: (props) => (
      <svg viewBox="0 0 45 45" {...props}>
        <path
          d="M22 10c10.5 1 16.5 8 16 29H15c0-9 10-6.5 8-21"
          strokeWidth="1.5"
          strokeLinecap="round"
        />
        <path
          d="M24 18c.38 2.32-4.68 1.97-5 4 0 0 .78-1.71-2.56-2.67C14.15 18.66 11 20 12 22c.38.75 3.12.83 2 4-2.58 2.34-4 2-4 8h12c1.65 0 3-1.35 3-3V18h-1z"
          strokeWidth="1.5"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>
    ),
    B: (props) => (
      <svg viewBox="0 0 45 45" {...props}>
        <g strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
          <path d="M9 36c3.39-.97 9.11-1.45 13.5-1.45 4.38 0 10.11.48 13.5 1.45V30H9v6zM15 30V16.5L22.5 8l7.5 8.5V30H15zM22.5 16l3.25-3.5L22.5 9l-3.25 3.5L22.5 16z" />
          <path d="M16 24.5h13" />
        </g>
      </svg>
    ),
    Q: (props) => (
      <svg viewBox="0 0 45 45" {...props}>
        <g strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
          <path d="M9 26c8.5-1.5 21-1.5 27 0l2-12-7 11V11l-5.5 13.5-3-15-3 15-5.5-13.5V25L9 14l2 12z" />
          <path d="M9 26c0 2 1.5 2 2.5 4 1 2.5 3 4.5 3 4.5h16s2-2 3-4.5c1-2 2.5-2 2.5-4-8.5-1.5-21-1.5-27 0z" />
          <path d="M11 38.5a35 35 1 0 0 0 23 0" />
        </g>
      </svg>
    ),
    K: (props) => (
      <svg viewBox="0 0 45 45" {...props}>
        <g strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
          <path d="M22.5 11.63V6M20 8h5" />
          <path d="M22.5 25s4.5-7.5 3-13.5c-6-1.5-9 1.5-9 6 .38 1.83 1.87 3.91 3 5" />
          <path d="M11.5 37c5.5 3.5 15.5 3.5 21 0v-7s9-4.5 6-10.5c-4-1-5 2.5-6 2.5-1.5-4.5-7.5-3.5-10.5-2S14.5 17 13 21.5c-1 0-2-3.5-6-2.5-3 6 6 10.5 6 10.5v7z" />
          <path d="M11.5 30c5.5-3 15.5-3 21 0" />
        </g>
      </svg>
    ),
  },
  Modern: {
    P: (props) => (
      <svg viewBox="0 0 45 45" {...props}>
        <circle cx="22.5" cy="18" r="6" />
        <path d="M13.5 36h18v-4l-9-10-9 10v4z" />
      </svg>
    ),
    R: (props) => (
      <svg viewBox="0 0 45 45" {...props}>
        <rect x="11" y="10" width="23" height="26" rx="2" />
        <path d="M11 16h23M16 10v6M29 10v6" strokeWidth="2" />
      </svg>
    ),
    N: (props) => (
      <svg viewBox="0 0 45 45" {...props}>
        <path d="M13 36h19L24 10l-6 4 4 6-8 6z" strokeLinejoin="round" />
      </svg>
    ),
    B: (props) => (
      <svg viewBox="0 0 45 45" {...props}>
        <path d="M22.5 8l-8 28h16z" />
        <circle cx="22.5" cy="18" r="3" fill="currentColor" className="opacity-40" />
      </svg>
    ),
    Q: (props) => (
      <svg viewBox="0 0 45 45" {...props}>
        <circle cx="22.5" cy="12" r="4" />
        <path d="M10.5 36h24l-4-20h-16z" />
      </svg>
    ),
    K: (props) => (
      <svg viewBox="0 0 45 45" {...props}>
        <rect x="18.5" y="8" width="8" height="28" />
        <path d="M12.5 36h20M14.5 16h16" strokeWidth="2" />
      </svg>
    ),
  },
  Minimal: {
    P: (props) => (
      <svg viewBox="0 0 45 45" {...props}>
        <path d="M22.5 35L15 20h15z" />
        <circle cx="22.5" cy="14" r="4" />
      </svg>
    ),
    R: (props) => (
      <svg viewBox="0 0 45 45" {...props}>
        <rect x="12" y="15" width="21" height="20" rx="2" />
        <path d="M12 15v-4h5v4h3v-4h5v4h3v-4h5v4" strokeLinecap="round" />
      </svg>
    ),
    N: (props) => (
      <svg viewBox="0 0 45 45" {...props}>
        <path d="M14 35h17V12l-8-2-9 5v8h5v5h-5z" />
      </svg>
    ),
    B: (props) => (
      <svg viewBox="0 0 45 45" {...props}>
        <path d="M22.5 8L34 35H11z" />
        <line x1="22.5" y1="8" x2="22.5" y2="35" stroke="currentColor" strokeOpacity="0.3" />
      </svg>
    ),
    Q: (props) => (
      <svg viewBox="0 0 45 45" {...props}>
        <circle cx="22.5" cy="22.5" r="12" />
        <path d="M22.5 6v6M39 22.5h-6M22.5 39v-6M6 22.5h6" strokeWidth="2" />
      </svg>
    ),
    K: (props) => (
      <svg viewBox="0 0 45 45" {...props}>
        <rect x="15" y="15" width="15" height="20" />
        <path d="M22.5 6v9M16 10h13" strokeWidth="2" />
      </svg>
    ),
  },
  Neo: {
    P: (props) => (
      <svg viewBox="0 0 45 45" {...props}>
        <circle cx="22.5" cy="15" r="4" strokeWidth="2" fill="none" />
        <path d="M22.5 20 v15 M15 35 h15" strokeWidth="2" fill="none" />
      </svg>
    ),
    R: (props) => (
      <svg viewBox="0 0 45 45" {...props}>
        <path d="M12 35h21V15H12z M12 15l-3-5h27l-3 5" strokeWidth="2" fill="none" />
      </svg>
    ),
    N: (props) => (
      <svg viewBox="0 0 45 45" {...props}>
        <path d="M15 35h15l-5-25-10 5 3 5-3 5-3 10z" strokeWidth="2" fill="none" />
      </svg>
    ),
    B: (props) => (
      <svg viewBox="0 0 45 45" {...props}>
        <path d="M22.5 5l10 30H12.5z" strokeWidth="2" fill="none" />
        <circle cx="22.5" cy="22" r="3" strokeWidth="1" fill="none" />
      </svg>
    ),
    Q: (props) => (
      <svg viewBox="0 0 45 45" {...props}>
        <rect x="12" y="12" width="21" height="21" transform="rotate(45 22.5 22.5)" strokeWidth="2" fill="none" />
        <circle cx="22.5" cy="22.5" r="3" fill="currentColor" />
      </svg>
    ),
    K: (props) => (
      <svg viewBox="0 0 45 45" {...props}>
        <path d="M15 35h15V15H15z M22.5 15V5 M17.5 10h10" strokeWidth="2" fill="none" />
      </svg>
    ),
  },
  Retro: {
    P: (props) => (
      <svg viewBox="0 0 45 45" {...props}>
        <path d="M20 15h5v5h-5zM18 20h9v10h-9zM15 30h15v5H15z" />
      </svg>
    ),
    R: (props) => (
      <svg viewBox="0 0 45 45" {...props}>
        <path d="M12 10h5v5h-5zM20 10h5v5h-5zM28 10h5v5h-5zM12 15h21v20H12z" />
      </svg>
    ),
    N: (props) => (
      <svg viewBox="0 0 45 45" {...props}>
        <path d="M20 10h10v5h-5v5h5v5H15v-5h-5v5H5v-5h5v-5h5zM15 30h15v5H15z" />
      </svg>
    ),
    B: (props) => (
      <svg viewBox="0 0 45 45" {...props}>
        <path d="M20 5h5v5h-5zM18 10h9v5h-9zM16 15h13v15H16zM13 30h19v5H13z" />
      </svg>
    ),
    Q: (props) => (
      <svg viewBox="0 0 45 45" {...props}>
        <path d="M10 10h5v5h-5zM30 10h5v5h-5zM18 5h9v5h-9zM15 15h15v15H15zM12 30h21v5H12z" />
      </svg>
    ),
    K: (props) => (
      <svg viewBox="0 0 45 45" {...props}>
        <path d="M20 5h5v10h5v-5h5v15H10v-15h5v5h5zM15 20h15v10H15zM12 30h21v5H12z" />
      </svg>
    ),
  },
};

const App = () => {
  const GAME_ID_STORAGE_KEY = 'go-chess:game-id';
  const PLAYER_TOKEN_STORAGE_KEY = 'go-chess:player-token';
  const PLAYER_COLOR_STORAGE_KEY = 'go-chess:player-color';
  const [board, setBoard] = useState(INITIAL_BOARD);
  const [turn, setTurn] = useState('white');
  const [selected, setSelected] = useState(null);
  const [validMoves, setValidMoves] = useState([]);
  const [winner, setWinner] = useState(null);
  const [check, setCheck] = useState(false);
  const [gameOverReason, setGameOverReason] = useState(null);

  // Menu & Modal States
  const [showResignConfirm, setShowResignConfirm] = useState(false);
  const [showGameMenu, setShowGameMenu] = useState(true); // Controls the New/Resume Game Popup
  const [joinId, setJoinId] = useState(''); // Input for joining game
  const [showSpectatePrompt, setShowSpectatePrompt] = useState(false);
  const [preferredColor, setPreferredColor] = useState('white');

  // Game Session State
  const [gameId, setGameId] = useState(null); // Current Active Game ID (null = local)
  const [playerToken, setPlayerToken] = useState(null);
  const [playerColor, setPlayerColor] = useState(null);
  const [isLoading, setIsLoading] = useState(false);
  const [validUciMoves, setValidUciMoves] = useState([]);
  const [lastFEN, setLastFEN] = useState(null);
  const isSpectating = Boolean(gameId && !playerToken);

  // Settings State
  const [darkMode, setDarkMode] = useState(true);
  const [boardTheme, setBoardTheme] = useState('wood');
  const [pieceTheme, setPieceTheme] = useState('Classic');
  const [showSettings, setShowSettings] = useState(false);

  // Audio refs
  const moveSound = useMemo(() => new Audio('https://assets.mixkit.co/active_storage/sfx/2070/2070-preview.mp3'), []);
  const captureSound = useMemo(() => new Audio('https://assets.mixkit.co/active_storage/sfx/2072/2072-preview.mp3'), []);

  // --- BACKEND SYNC ---
  const boardOrientation = playerColor === 'black' ? 'black' : 'white';
  const mapDisplayToBoard = (r, c) => (boardOrientation === 'black' ? { r: 7 - r, c: 7 - c } : { r, c });
  const mapBoardToDisplay = (r, c) => (boardOrientation === 'black' ? { r: 7 - r, c: 7 - c } : { r, c });

  const formatEndedBy = (endedBy) => {
    switch (endedBy) {
      case 'checkmate':
        return 'Checkmate';
      case 'stalemate':
        return 'Stalemate';
      case 'resignation':
        return 'Resigned';
      case 'draw_agreement':
        return 'Draw agreed';
      case 'draw_claim':
        return 'Draw claimed';
      case 'insufficient_material':
        return 'Insufficient material';
      case 'fifty_move':
        return 'Fifty-move rule';
      default:
        return null;
    }
  };

  const normalizeWinner = (winnerValue, resultValue) => {
    if (!winnerValue || winnerValue === 'none' || resultValue === 'draw' || resultValue === 'stalemate') {
      return 'Draw';
    }
    return winnerValue.charAt(0).toUpperCase() + winnerValue.slice(1);
  };

  const applyGameState = (data, options = {}) => {
    const { resetSelection = false } = options;
    const nextFEN = data.fen;
    const isSameFEN = nextFEN === lastFEN;

    if (!isSameFEN) {
      const { board: nextBoard, turn: nextTurn } = parseFEN(nextFEN);
      setBoard(nextBoard);
      setTurn(nextTurn);
      setLastFEN(nextFEN);
    }
    setCheck(Boolean(data.flags?.inCheck));
    if (data.playerColor) {
      setPlayerColor(data.playerColor);
      localStorage.setItem(PLAYER_COLOR_STORAGE_KEY, data.playerColor);
    } else if (!playerToken) {
      setPlayerColor(null);
      localStorage.removeItem(PLAYER_COLOR_STORAGE_KEY);
    }

    if (data.result && data.result !== 'ongoing') {
      setWinner(normalizeWinner(data.winner, data.result));
      setGameOverReason(formatEndedBy(data.endedBy) || data.result);
    } else {
      setWinner(null);
      setGameOverReason(null);
    }

    if (resetSelection) {
      setSelected(null);
      setValidMoves([]);
      setValidUciMoves([]);
    }
  };

  const fetchGame = async () => {
    const payload = await fetchJSON(`/api/v1/games/${encodeURIComponent(gameId)}`, {
      headers: playerToken ? { 'X-Player-Token': playerToken } : {},
    });
    applyGameState(payload);
  };

  // Fetch game state once for online games
  useEffect(() => {
    if (!gameId) return;

    let cancelled = false;
    const load = async () => {
      try {
        setIsLoading(true);
        await fetchGame();
        setShowGameMenu(false);
      } catch (e) {
        if (!cancelled) console.error('Fetch game failed:', e);
        if (!cancelled) {
          localStorage.removeItem(GAME_ID_STORAGE_KEY);
          localStorage.removeItem(PLAYER_TOKEN_STORAGE_KEY);
          localStorage.removeItem(PLAYER_COLOR_STORAGE_KEY);
          setGameId(null);
          setPlayerToken(null);
          setPlayerColor(null);
          setLastFEN(null);
        }
      } finally {
        if (!cancelled) setIsLoading(false);
      }
    };

    load();
    return () => {
      cancelled = true;
    };
  }, [gameId]);

  useEffect(() => {
    if (!gameId) return;
    const tokenParam = playerToken ? `?token=${encodeURIComponent(playerToken)}` : '';
    const streamUrl = `${API_BASE_URL}/api/v1/games/${encodeURIComponent(gameId)}/stream${tokenParam}`;
    const source = new EventSource(streamUrl);

    source.onmessage = (event) => {
      try {
        const payload = JSON.parse(event.data);
        applyGameState(payload, { resetSelection: true });
        if (payload.result && payload.result !== 'ongoing') {
          source.close();
        }
      } catch (e) {
        console.error('Stream parse failed:', e);
      }
    };

    source.onerror = (e) => {
      console.error('Stream error:', e);
    };

    return () => {
      source.close();
    };
  }, [gameId, playerToken]);

  useEffect(() => {
    const savedGameId = localStorage.getItem(GAME_ID_STORAGE_KEY);
    if (savedGameId) {
      setGameId(savedGameId);
    }
    const savedPlayerToken = localStorage.getItem(PLAYER_TOKEN_STORAGE_KEY);
    if (savedPlayerToken) {
      setPlayerToken(savedPlayerToken);
    }
    const savedPlayerColor = localStorage.getItem(PLAYER_COLOR_STORAGE_KEY);
    if (savedPlayerColor) {
      setPlayerColor(savedPlayerColor);
    }
  }, []);

  // --- GAME LOGIC ---

  useEffect(() => {
    if (gameId) return;
    // Check game state on every turn change (Runs on both local and synced updates)
    const inCheck = isKingInCheck(board, turn);
    setCheck(inCheck);

    if (inCheck && !winner) {
      let hasMoves = false;
      for (let r = 0; r < 8; r++) {
        for (let c = 0; c < 8; c++) {
          if (board[r][c] && getPieceColor(board[r][c]) === turn) {
            const moves = getValidMoves(board, r, c);
            if (moves.length > 0) {
              hasMoves = true;
              break;
            }
          }
        }
        if (hasMoves) break;
      }
      if (!hasMoves) {
        handleGameOver(turn === 'white' ? 'Black' : 'White', 'Checkmate');
      }
    }
  }, [turn, board, winner, gameId]);

  const handleGameOver = async (winnerName, reason) => {
    setWinner(winnerName);
    setGameOverReason(reason);
  };

  const handleSquareClick = async (r, c) => {
    if (winner) return;

    const boardPos = mapDisplayToBoard(r, c);

    if (gameId) {
      if (!playerToken) return;
      if (playerColor && playerColor !== turn) return;

      const piece = board[boardPos.r][boardPos.c];
      if (piece && getPieceColor(piece) === turn) {
        if (selected?.r === boardPos.r && selected?.c === boardPos.c) {
          setSelected(null);
          setValidMoves([]);
          setValidUciMoves([]);
        } else {
          setSelected({ r: boardPos.r, c: boardPos.c });
          setIsLoading(true);
          try {
            const from = squareFromCoords(boardPos.r, boardPos.c);
            const payload = await fetchJSON(
              `/api/v1/games/${encodeURIComponent(gameId)}/legal-moves?from=${encodeURIComponent(from)}`,
              { headers: playerToken ? { 'X-Player-Token': playerToken } : {} }
            );
            const uniqueTargets = new Map();
            payload.moves.forEach((uci) => {
              const target = uci.slice(2, 4);
              if (!uniqueTargets.has(target)) {
                uniqueTargets.set(target, coordsFromSquare(target));
              }
            });
            setValidMoves([...uniqueTargets.values()]);
            setValidUciMoves(payload.moves);
          } catch (e) {
            console.error('Legal moves fetch failed:', e);
            setValidMoves([]);
            setValidUciMoves([]);
          } finally {
            setIsLoading(false);
          }
        }
        return;
      }

      const move = validMoves.find((m) => m.r === boardPos.r && m.c === boardPos.c);
      if (selected && move) {
        const from = squareFromCoords(selected.r, selected.c);
        const to = squareFromCoords(boardPos.r, boardPos.c);
        const movingPiece = board[selected.r][selected.c];
        const isPromotion = movingPiece?.toLowerCase() === 'p' && (boardPos.r === 0 || boardPos.r === 7);
        const uci = `${from}${to}${isPromotion ? 'q' : ''}`;

        if (!validUciMoves.includes(uci)) return;

        const isCapture = board[boardPos.r][boardPos.c] !== null;
        setIsLoading(true);
        try {
          const payload = await fetchJSON(`/api/v1/games/${encodeURIComponent(gameId)}/moves`, {
            method: 'POST',
            body: JSON.stringify({ uci }),
            headers: playerToken ? { 'X-Player-Token': playerToken } : {},
          });
          const { board: nextBoard, turn: nextTurn } = parseFEN(payload.fen);
          setBoard(nextBoard);
          setTurn(nextTurn);
          setLastFEN(payload.fen);
          setCheck(Boolean(payload.flags?.inCheck));

          if (payload.result && payload.result !== 'ongoing') {
            setWinner(normalizeWinner(payload.winner, payload.result));
            setGameOverReason(formatEndedBy(payload.endedBy) || payload.result);
          } else {
            setWinner(null);
            setGameOverReason(null);
          }

          setSelected(null);
          setValidMoves([]);
          setValidUciMoves([]);

          if (isCapture) {
            captureSound.currentTime = 0;
            captureSound.play().catch(() => {});
          } else {
            moveSound.currentTime = 0;
            moveSound.play().catch(() => {});
          }
        } catch (e) {
          console.error('Move failed:', e);
        } finally {
          setIsLoading(false);
        }
      }
      return;
    }

    // Selection Logic
    if (board[boardPos.r][boardPos.c] && getPieceColor(board[boardPos.r][boardPos.c]) === turn) {
      if (selected?.r === boardPos.r && selected?.c === boardPos.c) {
        setSelected(null);
        setValidMoves([]);
      } else {
        setSelected({ r: boardPos.r, c: boardPos.c });
        setValidMoves(getValidMoves(board, boardPos.r, boardPos.c));
      }
      return;
    }

    // Move Logic
    const move = validMoves.find((m) => m.r === boardPos.r && m.c === boardPos.c);
    if (selected && move) {
      const isCapture = board[boardPos.r][boardPos.c] !== null;
      const newBoard = board.map((row) => [...row]);
      newBoard[boardPos.r][boardPos.c] = newBoard[selected.r][selected.c];
      newBoard[selected.r][selected.c] = null;

      // Promotion
      if (newBoard[boardPos.r][boardPos.c].toLowerCase() === 'p' && (boardPos.r === 0 || boardPos.r === 7)) {
        newBoard[boardPos.r][boardPos.c] = turn === 'white' ? 'Q' : 'q';
      }

      const nextTurn = turn === 'white' ? 'black' : 'white';

      // Optimistic Update (for instant feedback)
      setBoard(newBoard);
      setTurn(nextTurn);
      setSelected(null);
      setValidMoves([]);

      if (isCapture) {
        captureSound.currentTime = 0;
        captureSound.play().catch(() => {});
      } else {
        moveSound.currentTime = 0;
        moveSound.play().catch(() => {});
      }
    }
  };

  // --- MENU ACTIONS ---

  // 1. Start Fresh (Reset Local or trigger Menu)
  const resetLocalGame = () => {
    setBoard(INITIAL_BOARD);
    setTurn('white');
    setWinner(null);
    setCheck(false);
    setSelected(null);
    setValidMoves([]);
    setValidUciMoves([]);
    setGameOverReason(null);
    setGameId(null); // Clear ID -> Go Local
    setLastFEN(null);
    setPlayerToken(null);
    setPlayerColor(null);
    localStorage.removeItem(GAME_ID_STORAGE_KEY);
    localStorage.removeItem(PLAYER_TOKEN_STORAGE_KEY);
    localStorage.removeItem(PLAYER_COLOR_STORAGE_KEY);
    setShowResignConfirm(false);
    setShowGameMenu(false);
  };

  const exitSpectate = () => {
    setBoard(INITIAL_BOARD);
    setTurn('white');
    setWinner(null);
    setCheck(false);
    setSelected(null);
    setValidMoves([]);
    setValidUciMoves([]);
    setGameOverReason(null);
    setGameId(null);
    setLastFEN(null);
    setPlayerToken(null);
    setPlayerColor(null);
    localStorage.removeItem(GAME_ID_STORAGE_KEY);
    localStorage.removeItem(PLAYER_TOKEN_STORAGE_KEY);
    localStorage.removeItem(PLAYER_COLOR_STORAGE_KEY);
    setShowResignConfirm(false);
    setShowSpectatePrompt(false);
    setShowGameMenu(true);
  };

  // 2. Create Online Game
  const createOnlineGame = async () => {
    setIsLoading(true);
    try {
      const payload = await fetchJSON('/api/v1/games', {
        method: 'POST',
        body: JSON.stringify({ preferredColor }),
      });
      applyGameState(payload, { resetSelection: true });
      setGameId(payload.id);
      localStorage.setItem(GAME_ID_STORAGE_KEY, payload.id);
      setPlayerToken(payload.playerToken);
      localStorage.setItem(PLAYER_TOKEN_STORAGE_KEY, payload.playerToken);
      setPlayerColor(payload.playerColor);
      localStorage.setItem(PLAYER_COLOR_STORAGE_KEY, payload.playerColor);
      setShowGameMenu(false);
      setShowResignConfirm(false);
    } catch (e) {
      console.error('Error creating game', e);
    } finally {
      setIsLoading(false);
    }
  };

  // 3. Join Online Game
  const joinOnlineGame = async () => {
    if (joinId.length === 0) return;
    setIsLoading(true);
    try {
      const payload = await fetchJSON(`/api/v1/games/${encodeURIComponent(joinId)}/join`, {
        method: 'POST',
        body: JSON.stringify({}),
      });
      applyGameState(payload, { resetSelection: true });
      setGameId(payload.id);
      localStorage.setItem(GAME_ID_STORAGE_KEY, payload.id);
      setPlayerToken(payload.playerToken);
      localStorage.setItem(PLAYER_TOKEN_STORAGE_KEY, payload.playerToken);
      setPlayerColor(payload.playerColor);
      localStorage.setItem(PLAYER_COLOR_STORAGE_KEY, payload.playerColor);
      setShowGameMenu(false);
      setShowResignConfirm(false);
      setShowSpectatePrompt(false);
    } catch (e) {
      if (e.status === 409) {
        setShowSpectatePrompt(true);
        setShowGameMenu(false);
      } else {
        console.error('Error joining game', e);
      }
    } finally {
      setIsLoading(false);
    }
  };

  const spectateOnlineGame = async () => {
    if (joinId.length === 0) return;
    setIsLoading(true);
    try {
      const payload = await fetchJSON(`/api/v1/games/${encodeURIComponent(joinId)}`);
      applyGameState(payload, { resetSelection: true });
      setGameId(payload.id);
      localStorage.setItem(GAME_ID_STORAGE_KEY, payload.id);
      setPlayerToken(null);
      setPlayerColor(null);
      localStorage.removeItem(PLAYER_TOKEN_STORAGE_KEY);
      localStorage.removeItem(PLAYER_COLOR_STORAGE_KEY);
      setShowGameMenu(false);
      setShowResignConfirm(false);
      setShowSpectatePrompt(false);
    } catch (e) {
      console.error('Error spectating game', e);
    } finally {
      setIsLoading(false);
    }
  };

  // 4. Handle "New Game" Button Click
  const onNewGameClick = () => {
    if (winner) {
      // Game already over, just show menu
      setShowGameMenu(true);
    } else {
      // Game active, confirm resign first
      setShowResignConfirm(true);
    }
  };

  // 5. Confirm Resign (Triggers Menu after resigning)
  const confirmResignation = async () => {
    const opponent = turn === 'white' ? 'Black' : 'White';
    const reason = `${turn.charAt(0).toUpperCase() + turn.slice(1)} Resigned`;

    if (gameId) {
      setIsLoading(true);
      try {
        const payload = await fetchJSON(`/api/v1/games/${encodeURIComponent(gameId)}/resign`, {
          method: 'POST',
          body: JSON.stringify({ color: turn }),
          headers: playerToken ? { 'X-Player-Token': playerToken } : {},
        });
        setWinner(normalizeWinner(payload.winner, payload.result));
        setGameOverReason(formatEndedBy(payload.endedBy) || payload.result);
        setCheck(Boolean(payload.flags?.inCheck));
        setSelected(null);
        setValidMoves([]);
        setValidUciMoves([]);
      } catch (e) {
        console.error('Resign failed:', e);
      } finally {
        setIsLoading(false);
      }
    } else {
      // Update state locally
      await handleGameOver(opponent, reason);
    }

    setShowResignConfirm(false);
    // Removed setShowGameMenu(true) so the "Win/Resign" popup shows first.
    // The user will click "Next" on that popup to open the menu.
  };

  // --- RENDER HELPERS ---
  const getSquareColor = (r, c) => {
    const isDark = (r + c) % 2 === 1;
    return isDark ? BOARD_THEMES[boardTheme].dark : BOARD_THEMES[boardTheme].light;
  };

  const PieceComponent = ({ type, color }) => {
    const Component = Pieces[pieceTheme][type.toUpperCase()];
    const fillColor = color === 'white' ? '#ffffff' : '#272727';
    const strokeColor = color === 'white' ? '#272727' : '#ffffff';
    let finalFill = fillColor,
      finalStroke = strokeColor;

    if (pieceTheme === 'Retro') {
      finalFill = color === 'white' ? '#f1f5f9' : '#0f172a';
      finalStroke = 'none';
    } else if (pieceTheme === 'Modern') {
      finalStroke = color === 'white' ? '#1e293b' : '#e2e8f0';
    } else if (pieceTheme === 'Neo') {
      finalFill = 'none';
      finalStroke = color === 'white' ? '#ffffff' : '#272727';
    }

    return (
      <div className={`w-full h-full p-1.5 transition-transform duration-200 ${selected ? 'hover:scale-105' : ''}`}>
        <Component fill={finalFill} stroke={finalStroke} className="w-full h-full drop-shadow-sm" />
      </div>
    );
  };

  const copyGameId = () => {
    navigator.clipboard.writeText(gameId);
    // Could add toast here, but keeping simple
  };

  return (
    <div className={`nb-app min-h-screen w-full transition-colors duration-500 ${darkMode ? 'dark' : ''}`}>
      <div className="nb-shell max-w-6xl mx-auto p-4 md:p-8 flex flex-col md:flex-row gap-8 items-start justify-center">
        {/* --- LEFT PANEL --- */}
        <div className="flex-1 w-full max-w-[600px] flex flex-col gap-4">
          <div className="flex justify-between items-center mb-2">
            <h1 className="nb-title text-3xl md:text-4xl flex items-center gap-3">
              <Ghost className="w-6 h-6" /> Chess
            </h1>

            <button
              onClick={() => setShowSettings(!showSettings)}
              className={`nb-icon-button ${showSettings ? 'is-active' : ''}`}
            >
              <Settings className="w-5 h-5" />
            </button>
          </div>

          <div
            className={`relative aspect-square w-full nb-board-frame transition-colors duration-300 ${BOARD_THEMES[boardTheme].border}`}
          >
            {/* Loading Overlay */}
            {isLoading && (
              <div className="nb-overlay absolute inset-0 z-50 flex items-center justify-center">
                <RefreshCw className="w-10 h-10 text-white animate-spin" />
              </div>
            )}

            {/* Game Over Overlay */}
            {winner && !showGameMenu && (
              <div className="nb-overlay nb-overlay--dark absolute inset-0 z-40 flex flex-col items-center justify-center animate-in fade-in">
                <Trophy className="w-16 h-16 text-yellow-300 mb-4 animate-bounce" />
                <h2 className="nb-title text-4xl text-white mb-1">
                  {winner === 'Draw' ? 'Draw' : `${winner} Wins!`}
                </h2>
                {gameOverReason && <p className="text-lg text-white/80 font-medium mb-4">{gameOverReason}</p>}

                <button
                  onClick={() => setShowGameMenu(true)}
                  className="nb-button nb-button--primary mt-6 flex items-center gap-2"
                >
                  Next <ArrowRight className="w-5 h-5" />
                </button>
              </div>
            )}

            {/* Resign Confirmation Overlay */}
            {showResignConfirm && !winner && (
              <div className="nb-overlay nb-overlay--dark absolute inset-0 z-50 flex flex-col items-center justify-center animate-in fade-in p-6 text-center">
                <div className="nb-modal p-8 max-w-sm w-full mx-4">
                  <div className="nb-icon-badge nb-icon-badge--danger mx-auto mb-4">
                    <AlertTriangle className="w-8 h-8" />
                  </div>
                  <h3 className="nb-title text-2xl mb-2">Resign Game?</h3>
                  <p className="nb-muted mb-8">To start a new game or join another, you must resign the current match.</p>
                  <div className="flex gap-3 justify-center">
                    <button
                      onClick={() => setShowResignConfirm(false)}
                      className="nb-button nb-button--neutral w-full"
                    >
                      Cancel
                    </button>
                    <button
                      onClick={confirmResignation}
                      className="nb-button nb-button--danger w-full"
                    >
                      Resign
                    </button>
                  </div>
                </div>
              </div>
            )}

            {/* Spectate Prompt Overlay */}
            {showSpectatePrompt && (
              <div className="nb-overlay nb-overlay--dark absolute inset-0 z-50 flex flex-col items-center justify-center animate-in fade-in p-6 text-center">
                <div className="nb-modal p-8 max-w-sm w-full mx-4">
                  <div className="nb-icon-badge nb-icon-badge--primary mx-auto mb-4">
                    <Users className="w-8 h-8" />
                  </div>
                  <h3 className="nb-title text-2xl mb-2">Room is full</h3>
                  <p className="nb-muted mb-8">You can still spectate this match.</p>
                  <div className="flex gap-3 justify-center">
                    <button
                      onClick={() => {
                        setShowSpectatePrompt(false);
                        setShowGameMenu(true);
                      }}
                      className="nb-button nb-button--neutral w-full"
                    >
                      Cancel
                    </button>
                    <button
                      onClick={spectateOnlineGame}
                      className="nb-button nb-button--primary w-full"
                    >
                      Spectate
                    </button>
                  </div>
                </div>
              </div>
            )}

            {/* Game Menu (Create/Join/Local) Overlay */}
            {showGameMenu && (
              <div className="nb-overlay nb-overlay--dark absolute inset-0 z-50 flex flex-col items-center justify-center animate-in fade-in p-6 text-center">
                <div className="nb-modal p-6 md:p-8 max-w-md w-full">
                  <h2 className="nb-title text-3xl mb-2">New Game</h2>
                  <p className="nb-muted mb-8">Choose how you want to play</p>

                  <div className="space-y-3">
                    {/* Option 1: Play Local */}
                    <button onClick={resetLocalGame} className="nb-option nb-option--primary w-full flex items-center group">
                      <div className="nb-icon-badge nb-icon-badge--primary mr-4">
                        <Users className="w-6 h-6" />
                      </div>
                      <div className="text-left">
                        <h3 className="nb-option-title">Play Local</h3>
                        <p className="nb-muted text-xs">Pass and play on this device</p>
                      </div>
                    </button>

                    <div className="nb-card nb-card--soft p-4">
                      <h3 className="nb-label text-left mb-2 flex items-center gap-2">
                        <Trophy className="w-4 h-4" /> Play As
                      </h3>
                      <div className="grid grid-cols-2 gap-2">
                        <button
                          onClick={() => setPreferredColor('white')}
                          className={`nb-toggle ${preferredColor === 'white' ? 'is-active' : ''}`}
                        >
                          White
                        </button>
                        <button
                          onClick={() => setPreferredColor('black')}
                          className={`nb-toggle ${preferredColor === 'black' ? 'is-active' : ''}`}
                        >
                          Black
                        </button>
                      </div>
                    </div>

                    {/* Option 2: Create Online */}
                    <button onClick={createOnlineGame} className="nb-option nb-option--success w-full flex items-center group">
                      <div className="nb-icon-badge nb-icon-badge--success mr-4">
                        <Globe className="w-6 h-6" />
                      </div>
                      <div className="text-left">
                        <h3 className="nb-option-title">Create Online Game</h3>
                        <p className="nb-muted text-xs">Get a Game ID to share with a friend</p>
                      </div>
                    </button>

                    {/* Option 3: Join Online */}
                    <div className="nb-card nb-card--soft p-4">
                      <h3 className="nb-label text-left mb-2 flex items-center gap-2">
                        <Play className="w-4 h-4" /> Join Game
                      </h3>
                      <div className="flex gap-2">
                        <input
                          type="text"
                          placeholder="Enter Game ID"
                          value={joinId}
                          onChange={(e) => setJoinId(e.target.value)}
                          className="nb-input flex-1 font-mono"
                        />
                        <button
                          onClick={joinOnlineGame}
                          disabled={!joinId}
                          className="nb-button nb-button--primary"
                        >
                          Join
                        </button>
                      </div>
                    </div>
                  </div>

                  <button
                    onClick={() => setShowGameMenu(false)}
                    className="nb-link mt-6"
                  >
                    Cancel
                  </button>
                </div>
              </div>
            )}

            <div className="w-full h-full grid grid-cols-8 grid-rows-8">
              {Array.from({ length: 8 }, (_, displayR) =>
                Array.from({ length: 8 }, (_, displayC) => {
                  const boardPos = mapDisplayToBoard(displayR, displayC);
                  const piece = board[boardPos.r][boardPos.c];
                  const isValid = validMoves.some((m) => m.r === boardPos.r && m.c === boardPos.c);
                  const isSelected = selected?.r === boardPos.r && selected?.c === boardPos.c;
                  const isCheckSquare =
                    check && piece && getPieceColor(piece) === turn && piece.toLowerCase() === 'k';

                  return (
                    <div
                      key={`${displayR}-${displayC}`}
                      onClick={() => handleSquareClick(displayR, displayC)}
                      style={{ backgroundColor: getSquareColor(boardPos.r, boardPos.c) }}
                      className="nb-square relative flex items-center justify-center cursor-pointer"
                    >
                      {/* Rank/File Indicators */}
                      {displayC === 0 && displayR === 7 && (
                        <span
                          className={`absolute bottom-0.5 left-1 text-[10px] font-bold ${
                            getPieceColor(piece) === 'white' ? 'text-slate-800' : 'opacity-60'
                          }`}
                        >
                          a1
                        </span>
                      )}

                      {isSelected && <div className="absolute inset-0 nb-square-selected" />}
                      {isValid && !piece && <div className="nb-move-dot" />}
                      {isValid && piece && (
                        <div className="absolute inset-0">
                          <div className="nb-corner nb-corner--tl" />
                          <div className="nb-corner nb-corner--tr" />
                          <div className="nb-corner nb-corner--bl" />
                          <div className="nb-corner nb-corner--br" />
                        </div>
                      )}
                      {isCheckSquare && (
                        <div
                          className="absolute inset-0"
                          style={{
                            background: 'radial-gradient(circle, rgba(255,0,0,0.8) 0%, rgba(255,0,0,0) 70%)',
                          }}
                        />
                      )}
                      {piece && (
                        <div className="w-4/5 h-4/5 z-10">
                          <PieceComponent type={piece} color={getPieceColor(piece)} />
                        </div>
                      )}
                    </div>
                  );
                })
              )}
            </div>
          </div>

          <div className="nb-status-bar p-4 rounded-xl flex flex-col sm:flex-row justify-between items-center gap-3">
            <div className="flex items-center gap-3">
              <div
                className={`nb-turn-dot ${turn === 'white' ? 'is-white' : 'is-black'}`}
              ></div>
              <span className="font-semibold capitalize">{turn}'s Turn</span>
              {isSpectating && (
                <span className="nb-chip">
                  <Eye className="w-3.5 h-3.5" /> Spectating
                </span>
              )}
              {check && !winner && <span className="nb-check ml-2">CHECK!</span>}
            </div>

            {/* Game ID Display */}
            {!showGameMenu && (
              <div className="nb-id flex items-center gap-2 px-3 py-1.5 rounded-lg">
                <span className="nb-label text-[11px]">ID:</span>
                {gameId ? (
                  <>
                    <span className="font-mono nb-id-value tracking-wider">{gameId}</span>
                    <button onClick={copyGameId} className="nb-icon-button nb-icon-button--tiny">
                      <Copy className="w-3 h-3" />
                    </button>
                  </>
                ) : (
                  <span className="text-xs font-semibold nb-muted">- offline -</span>
                )}
              </div>
            )}

            {isSpectating ? (
              <button
                onClick={exitSpectate}
                className="nb-button nb-button--primary nb-button--compact flex items-center gap-1"
              >
                <ArrowRight className="w-4 h-4" /> Exit
              </button>
            ) : (
              <button
                onClick={onNewGameClick}
                className="nb-button nb-button--danger nb-button--compact flex items-center gap-1"
              >
                <Flag className="w-4 h-4" /> New Game
              </button>
            )}
          </div>
        </div>

        {/* --- RIGHT PANEL --- */}
        {showSettings && (
          <div className="nb-panel w-full md:w-80 flex flex-col gap-6 animate-in slide-in-from-right-10 fade-in duration-300">
            <div className="flex items-center justify-between mb-2">
              <h2 className="nb-title text-xl">Settings</h2>
              <button
                onClick={() => setShowSettings(false)}
                className="nb-icon-button nb-icon-button--danger"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="nb-card p-6 space-y-6">
              <div className="flex items-center justify-between">
                <label className="nb-label flex items-center gap-2">
                  App Theme
                </label>
                <button
                  onClick={() => setDarkMode(!darkMode)}
                  className={`nb-icon-button nb-icon-button--toggle ${darkMode ? 'is-active' : ''}`}
                >
                  {darkMode ? <Sun className="w-5 h-5" /> : <Moon className="w-5 h-5" />}
                </button>
              </div>
              <hr className="nb-divider" />
              <div className="space-y-3">
                <label className="nb-label flex items-center gap-2">
                  <Palette className="w-4 h-4" /> Board Theme
                </label>
                <div className="grid grid-cols-2 gap-2">
                  {Object.entries(BOARD_THEMES).map(([key, theme]) => (
                    <button
                      key={key}
                      onClick={() => setBoardTheme(key)}
                      className={`nb-swatch ${boardTheme === key ? 'is-active' : ''}`}
                    >
                      <div className="absolute inset-0 flex">
                        <div className="w-1/2 h-full" style={{ backgroundColor: theme.light }} />
                        <div className="w-1/2 h-full" style={{ backgroundColor: theme.dark }} />
                      </div>
                      <span className="nb-swatch-label">
                        {theme.name}
                      </span>
                    </button>
                  ))}
                </div>
              </div>
              <hr className="nb-divider" />
              <div className="space-y-3">
                <label className="nb-label flex items-center gap-2">
                  <LayoutGrid className="w-4 h-4" /> Piece Style
                </label>
                <div className="flex flex-col gap-2">
                  {Object.keys(Pieces).map((key) => (
                    <button
                      key={key}
                      onClick={() => setPieceTheme(key)}
                      className={`nb-choice flex items-center gap-3 ${pieceTheme === key ? 'is-active' : ''}`}
                    >
                      <div className="w-8 h-8 text-slate-800 dark:text-slate-200">
                        {React.createElement(Pieces[key].N, {
                          stroke: key === 'Neo' ? 'currentColor' : undefined,
                          fill: key === 'Neo' ? 'none' : 'currentColor',
                        })}
                      </div>
                      <span className="font-medium text-sm">{key}</span>
                    </button>
                  ))}
                </div>
              </div>
            </div>
            <div className="nb-footnote mt-auto">
              Lite version. En Passant & Castling coming soon.
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default App;
