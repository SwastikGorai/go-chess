import React, { useState, useEffect, useMemo, useRef } from 'react';
import { 
  Settings, 
  RefreshCw, 
  Moon, 
  Sun, 
  Trophy, 
  RotateCcw, 
  Palette, 
  LayoutGrid, 
  Ghost,
  X,
  Flag,
  AlertTriangle,
  Globe,
  Copy,
  Users,
  Play,
  ArrowRight
} from 'lucide-react';
import { initializeApp } from 'firebase/app';
import { getAuth, signInAnonymously, onAuthStateChanged, signInWithCustomToken } from 'firebase/auth';
import { getFirestore, doc, setDoc, getDoc, onSnapshot, updateDoc } from 'firebase/firestore';

// --- FIREBASE SETUP ---
const firebaseConfig = JSON.parse(__firebase_config);
const app = initializeApp(firebaseConfig);
const auth = getAuth(app);
const db = getFirestore(app);
const appId = typeof __app_id !== 'undefined' ? __app_id : 'default-app-id';

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
const getPieceColor = (piece) => piece ? (isWhite(piece) ? 'white' : 'black') : null;

const getValidMoves = (board, row, col, checkSafety = true) => {
  const piece = board[row][col];
  if (!piece) return [];
  
  const moves = [];
  const color = getPieceColor(piece);
  
  const directions = {
    pawn: [],
    rook: [[0, 1], [0, -1], [1, 0], [-1, 0]],
    bishop: [[1, 1], [1, -1], [-1, 1], [-1, -1]],
    knight: [[2, 1], [2, -1], [-2, 1], [-2, -1], [1, 2], [1, -2], [-1, 2], [-1, -2]],
    queen: [[0, 1], [0, -1], [1, 0], [-1, 0], [1, 1], [1, -1], [-1, 1], [-1, -1]],
    king: [[0, 1], [0, -1], [1, 0], [-1, 0], [1, 1], [1, -1], [-1, 1], [-1, -1]]
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
    [[dir, 1], [dir, -1]].forEach(([dr, dc]) => {
      const r = row + dr, c = col + dc;
      if (board[r]?.[c] && getPieceColor(board[r][c]) !== color) {
        moves.push({ r, c });
      }
    });
  } else if (type === 'n' || type === 'k') {
    (type === 'n' ? directions.knight : directions.king).forEach(([dr, dc]) => {
      const r = row + dr, c = col + dc;
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
      let r = row + dr, c = col + dc;
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
    return moves.filter(move => {
      const newBoard = board.map(r => [...r]);
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
        if (moves.some(m => m.r === kingPos.r && m.c === kingPos.c)) {
          return true;
        }
      }
    }
  }
  return false;
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
    P: (props) => <svg viewBox="0 0 45 45" {...props}><path d="M22.5 9c-2.21 0-4 1.79-4 4 0 .89.29 1.71.78 2.38C17.33 16.5 16 18.59 16 21c0 2.03.94 3.84 2.41 5.03-3 1.06-7.41 5.55-7.41 13.47h23c0-7.92-4.41-12.41-7.41-13.47 1.47-1.19 2.41-3 2.41-5.03 0-2.41-1.33-4.5-3.28-5.62.49-.67.78-1.49.78-2.38 0-2.21-1.79-4-4-4z" strokeWidth="1.5" strokeLinecap="round" /></svg>,
    R: (props) => <svg viewBox="0 0 45 45" {...props}><path d="M9 39h27v-3H9v3zM12 36v-4h21v4H12zM11 14V9h4v2h5V9h5v2h5V9h4v5h-2.932L22.5 19.5 13.932 14H11zM12 32h21V19.5L22.5 30 12 19.5V32z" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" /></svg>,
    N: (props) => <svg viewBox="0 0 45 45" {...props}><path d="M22 10c10.5 1 16.5 8 16 29H15c0-9 10-6.5 8-21" strokeWidth="1.5" strokeLinecap="round" /><path d="M24 18c.38 2.32-4.68 1.97-5 4 0 0 .78-1.71-2.56-2.67C14.15 18.66 11 20 12 22c.38.75 3.12.83 2 4-2.58 2.34-4 2-4 8h12c1.65 0 3-1.35 3-3V18h-1z" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" /></svg>,
    B: (props) => <svg viewBox="0 0 45 45" {...props}><g strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"><path d="M9 36c3.39-.97 9.11-1.45 13.5-1.45 4.38 0 10.11.48 13.5 1.45V30H9v6zM15 30V16.5L22.5 8l7.5 8.5V30H15zM22.5 16l3.25-3.5L22.5 9l-3.25 3.5L22.5 16z" /><path d="M16 24.5h13" /></g></svg>,
    Q: (props) => <svg viewBox="0 0 45 45" {...props}><g strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"><path d="M9 26c8.5-1.5 21-1.5 27 0l2-12-7 11V11l-5.5 13.5-3-15-3 15-5.5-13.5V25L9 14l2 12z" /><path d="M9 26c0 2 1.5 2 2.5 4 1 2.5 3 4.5 3 4.5h16s2-2 3-4.5c1-2 2.5-2 2.5-4-8.5-1.5-21-1.5-27 0z" /><path d="M11 38.5a35 35 1 0 0 0 23 0" /></g></svg>,
    K: (props) => <svg viewBox="0 0 45 45" {...props}><g strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"><path d="M22.5 11.63V6M20 8h5" /><path d="M22.5 25s4.5-7.5 3-13.5c-6-1.5-9 1.5-9 6 .38 1.83 1.87 3.91 3 5" /><path d="M11.5 37c5.5 3.5 15.5 3.5 21 0v-7s9-4.5 6-10.5c-4-1-5 2.5-6 2.5-1.5-4.5-7.5-3.5-10.5-2S14.5 17 13 21.5c-1 0-2-3.5-6-2.5-3 6 6 10.5 6 10.5v7z" /><path d="M11.5 30c5.5-3 15.5-3 21 0" /></g></svg>,
  },
  Modern: {
    P: (props) => <svg viewBox="0 0 45 45" {...props}><circle cx="22.5" cy="18" r="6" /><path d="M13.5 36h18v-4l-9-10-9 10v4z" /></svg>,
    R: (props) => <svg viewBox="0 0 45 45" {...props}><rect x="11" y="10" width="23" height="26" rx="2" /><path d="M11 16h23M16 10v6M29 10v6" strokeWidth="2" /></svg>,
    N: (props) => <svg viewBox="0 0 45 45" {...props}><path d="M13 36h19L24 10l-6 4 4 6-8 6z" strokeLinejoin="round" /></svg>,
    B: (props) => <svg viewBox="0 0 45 45" {...props}><path d="M22.5 8l-8 28h16z" /><circle cx="22.5" cy="18" r="3" fill="currentColor" className="opacity-40" /></svg>,
    Q: (props) => <svg viewBox="0 0 45 45" {...props}><circle cx="22.5" cy="12" r="4" /><path d="M10.5 36h24l-4-20h-16z" /></svg>,
    K: (props) => <svg viewBox="0 0 45 45" {...props}><rect x="18.5" y="8" width="8" height="28" /><path d="M12.5 36h20M14.5 16h16" strokeWidth="2" /></svg>,
  },
  Minimal: {
    P: (props) => <svg viewBox="0 0 45 45" {...props}><path d="M22.5 35L15 20h15z" /><circle cx="22.5" cy="14" r="4" /></svg>,
    R: (props) => <svg viewBox="0 0 45 45" {...props}><rect x="12" y="15" width="21" height="20" rx="2" /><path d="M12 15v-4h5v4h3v-4h5v4h3v-4h5v4" strokeLinecap="round"/></svg>,
    N: (props) => <svg viewBox="0 0 45 45" {...props}><path d="M14 35h17V12l-8-2-9 5v8h5v5h-5z" /></svg>,
    B: (props) => <svg viewBox="0 0 45 45" {...props}><path d="M22.5 8L34 35H11z" /><line x1="22.5" y1="8" x2="22.5" y2="35" stroke="currentColor" strokeOpacity="0.3" /></svg>,
    Q: (props) => <svg viewBox="0 0 45 45" {...props}><circle cx="22.5" cy="22.5" r="12" /><path d="M22.5 6v6M39 22.5h-6M22.5 39v-6M6 22.5h6" strokeWidth="2"/></svg>,
    K: (props) => <svg viewBox="0 0 45 45" {...props}><rect x="15" y="15" width="15" height="20" /><path d="M22.5 6v9M16 10h13" strokeWidth="2"/></svg>,
  },
  Neo: { 
    P: (props) => <svg viewBox="0 0 45 45" {...props}><circle cx="22.5" cy="15" r="4" strokeWidth="2" fill="none"/><path d="M22.5 20 v15 M15 35 h15" strokeWidth="2" fill="none"/></svg>,
    R: (props) => <svg viewBox="0 0 45 45" {...props}><path d="M12 35h21V15H12z M12 15l-3-5h27l-3 5" strokeWidth="2" fill="none"/></svg>,
    N: (props) => <svg viewBox="0 0 45 45" {...props}><path d="M15 35h15l-5-25-10 5 3 5-3 5-3 10z" strokeWidth="2" fill="none"/></svg>,
    B: (props) => <svg viewBox="0 0 45 45" {...props}><path d="M22.5 5l10 30H12.5z" strokeWidth="2" fill="none"/><circle cx="22.5" cy="22" r="3" strokeWidth="1" fill="none"/></svg>,
    Q: (props) => <svg viewBox="0 0 45 45" {...props}><rect x="12" y="12" width="21" height="21" transform="rotate(45 22.5 22.5)" strokeWidth="2" fill="none"/><circle cx="22.5" cy="22.5" r="3" fill="currentColor"/></svg>,
    K: (props) => <svg viewBox="0 0 45 45" {...props}><path d="M15 35h15V15H15z M22.5 15V5 M17.5 10h10" strokeWidth="2" fill="none"/></svg>,
  },
  Retro: {
    P: (props) => <svg viewBox="0 0 45 45" {...props}><path d="M20 15h5v5h-5zM18 20h9v10h-9zM15 30h15v5H15z" /></svg>,
    R: (props) => <svg viewBox="0 0 45 45" {...props}><path d="M12 10h5v5h-5zM20 10h5v5h-5zM28 10h5v5h-5zM12 15h21v20H12z" /></svg>,
    N: (props) => <svg viewBox="0 0 45 45" {...props}><path d="M20 10h10v5h-5v5h5v5H15v-5h-5v5H5v-5h5v-5h5zM15 30h15v5H15z" /></svg>,
    B: (props) => <svg viewBox="0 0 45 45" {...props}><path d="M20 5h5v5h-5zM18 10h9v5h-9zM16 15h13v15H16zM13 30h19v5H13z" /></svg>,
    Q: (props) => <svg viewBox="0 0 45 45" {...props}><path d="M10 10h5v5h-5zM30 10h5v5h-5zM18 5h9v5h-9zM15 15h15v15H15zM12 30h21v5H12z" /></svg>,
    K: (props) => <svg viewBox="0 0 45 45" {...props}><path d="M20 5h5v10h5v-5h5v15H10v-15h5v5h5zM15 20h15v10H15zM12 30h21v5H12z" /></svg>,
  }
};

const App = () => {
  const [board, setBoard] = useState(INITIAL_BOARD);
  const [turn, setTurn] = useState('white');
  const [selected, setSelected] = useState(null);
  const [validMoves, setValidMoves] = useState([]);
  const [winner, setWinner] = useState(null);
  const [check, setCheck] = useState(false);
  const [gameOverReason, setGameOverReason] = useState(null);
  
  // Menu & Modal States
  const [showResignConfirm, setShowResignConfirm] = useState(false);
  const [showGameMenu, setShowGameMenu] = useState(false); // Controls the New/Resume Game Popup
  const [joinId, setJoinId] = useState(''); // Input for joining game
  
  // Game Session State
  const [user, setUser] = useState(null);
  const [gameId, setGameId] = useState(null); // Current Active Game ID (null = local)
  const [isLoading, setIsLoading] = useState(false);
  
  // Settings State
  const [darkMode, setDarkMode] = useState(true);
  const [boardTheme, setBoardTheme] = useState('wood');
  const [pieceTheme, setPieceTheme] = useState('Classic');
  const [showSettings, setShowSettings] = useState(false);

  // Audio refs
  const moveSound = useMemo(() => new Audio('https://assets.mixkit.co/active_storage/sfx/2070/2070-preview.mp3'), []);
  const captureSound = useMemo(() => new Audio('https://assets.mixkit.co/active_storage/sfx/2072/2072-preview.mp3'), []);

  // --- FIREBASE AUTH & SYNC ---
  
  // 1. Init Auth
  useEffect(() => {
    const initAuth = async () => {
      if (typeof __initial_auth_token !== 'undefined' && __initial_auth_token) {
        try {
          await signInWithCustomToken(auth, __initial_auth_token);
        } catch (e) {
          console.error("Custom token sign-in failed, falling back to anon:", e);
          await signInAnonymously(auth);
        }
      } else {
        await signInAnonymously(auth);
      }
    };
    initAuth();
    const unsubscribe = onAuthStateChanged(auth, setUser);
    return () => unsubscribe();
  }, []);

  // 2. Sync Game Data when GameID changes
  useEffect(() => {
    if (!gameId || !user) return;

    setIsLoading(true);
    const gameRef = doc(db, 'artifacts', appId, 'public', 'data', 'games', gameId);

    const unsubscribe = onSnapshot(gameRef, (docSnap) => {
      setIsLoading(false);
      if (docSnap.exists()) {
        const data = docSnap.data();
        // Parse board from JSON string (Firestore doesn't support nested arrays natively well)
        try {
            setBoard(JSON.parse(data.board));
        } catch(e) { console.error("Board parse error", e); }
        
        setTurn(data.turn);
        setWinner(data.winner || null);
        setGameOverReason(data.gameOverReason || null);
      } else {
        // Game doesn't exist? Maybe go back to local
        console.log("Game not found");
      }
    }, (error) => {
        console.error("Snapshot error:", error);
        setIsLoading(false);
    });

    return () => unsubscribe();
  }, [gameId, user]);

  // --- GAME LOGIC ---

  useEffect(() => {
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
  }, [turn, board, winner]);

  const handleGameOver = async (winnerName, reason) => {
    setWinner(winnerName);
    setGameOverReason(reason);
    
    // If online, sync game over state
    if (gameId && user) {
        try {
            const gameRef = doc(db, 'artifacts', appId, 'public', 'data', 'games', gameId);
            await updateDoc(gameRef, {
                winner: winnerName,
                gameOverReason: reason
            });
        } catch(e) { console.error("Error updating winner", e); }
    }
  };

  const handleSquareClick = async (r, c) => {
    if (winner) return;

    // Selection Logic
    if (board[r][c] && getPieceColor(board[r][c]) === turn) {
      if (selected?.r === r && selected?.c === c) {
        setSelected(null);
        setValidMoves([]);
      } else {
        setSelected({ r, c });
        setValidMoves(getValidMoves(board, r, c));
      }
      return;
    }

    // Move Logic
    const move = validMoves.find(m => m.r === r && m.c === c);
    if (selected && move) {
      const isCapture = board[r][c] !== null;
      const newBoard = board.map(row => [...row]);
      newBoard[r][c] = newBoard[selected.r][selected.c];
      newBoard[selected.r][selected.c] = null;
      
      // Promotion
      if (newBoard[r][c].toLowerCase() === 'p' && (r === 0 || r === 7)) {
        newBoard[r][c] = turn === 'white' ? 'Q' : 'q';
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

      // Sync to Firebase if Online
      if (gameId && user) {
          try {
             const gameRef = doc(db, 'artifacts', appId, 'public', 'data', 'games', gameId);
             await updateDoc(gameRef, {
                 board: JSON.stringify(newBoard),
                 turn: nextTurn
             });
          } catch(e) { console.error("Error syncing move", e); }
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
    setGameOverReason(null);
    setGameId(null); // Clear ID -> Go Local
    setShowResignConfirm(false);
    setShowGameMenu(false);
  };

  // 2. Create Online Game
  const createOnlineGame = async () => {
    if (!user) return;
    setIsLoading(true);
    const newId = Math.random().toString(36).substring(2, 8).toUpperCase();
    
    const gameRef = doc(db, 'artifacts', appId, 'public', 'data', 'games', newId);
    
    // Reset local state first
    setBoard(INITIAL_BOARD);
    setTurn('white');
    setWinner(null);
    setCheck(false);
    setGameOverReason(null);
    setSelected(null);
    setValidMoves([]);

    try {
        await setDoc(gameRef, {
            board: JSON.stringify(INITIAL_BOARD),
            turn: 'white',
            winner: null,
            gameOverReason: null,
            createdAt: new Date().toISOString()
        });
        setGameId(newId);
        setShowGameMenu(false);
        setShowResignConfirm(false);
    } catch (e) {
        console.error("Error creating game", e);
    } finally {
        setIsLoading(false);
    }
  };

  // 3. Join Online Game
  const joinOnlineGame = () => {
      if (joinId.length > 0) {
          setGameId(joinId); // This triggers the useEffect sync
          setShowGameMenu(false);
          setShowResignConfirm(false);
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
      
      // Update state locally and remotely
      await handleGameOver(opponent, reason);
      
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
    let finalFill = fillColor, finalStroke = strokeColor;
    
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
    <div className={`min-h-screen w-full transition-colors duration-500 font-sans ${darkMode ? 'dark bg-slate-900 text-slate-100' : 'bg-slate-50 text-slate-900'}`}>
      
      <div className="max-w-6xl mx-auto p-4 md:p-8 flex flex-col md:flex-row gap-8 items-start justify-center">
        
        {/* --- LEFT PANEL --- */}
        <div className="flex-1 w-full max-w-[600px] flex flex-col gap-4">
          
          <div className="flex justify-between items-center mb-2">
            <h1 className="text-2xl font-bold tracking-tight flex items-center gap-2">
              <Ghost className="w-6 h-6" /> Chess
            </h1>
            
            <button 
              onClick={() => setShowSettings(!showSettings)} 
              className={`
                w-10 h-10 flex items-center justify-center rounded-full transition-all duration-300 focus:outline-none focus:ring-2 focus:ring-blue-500
                ${showSettings 
                  ? 'bg-blue-100 dark:bg-blue-900 text-blue-600 dark:text-blue-400 rotate-90' 
                  : 'bg-slate-200 dark:bg-slate-800 text-slate-700 dark:text-slate-200 hover:bg-slate-300 dark:hover:bg-slate-700'
                }
              `}
            >
              <Settings className="w-5 h-5" />
            </button>
          </div>

          <div className={`relative aspect-square w-full rounded-lg overflow-hidden shadow-2xl border-[12px] transition-colors duration-300 ${BOARD_THEMES[boardTheme].border}`}>
            
            {/* Loading Overlay */}
            {isLoading && (
                 <div className="absolute inset-0 z-50 bg-black/50 flex items-center justify-center">
                     <RefreshCw className="w-10 h-10 text-white animate-spin" />
                 </div>
            )}

            {/* Game Over Overlay */}
            {winner && !showGameMenu && (
              <div className="absolute inset-0 z-40 bg-black/60 flex flex-col items-center justify-center backdrop-blur-sm animate-in fade-in">
                <Trophy className="w-16 h-16 text-yellow-400 mb-4 animate-bounce" />
                <h2 className="text-4xl font-bold text-white mb-1">{winner} Wins!</h2>
                {gameOverReason && <p className="text-lg text-white/80 font-medium mb-4">{gameOverReason}</p>}
                
                <button 
                  onClick={() => setShowGameMenu(true)}
                  className="mt-6 px-8 py-3 bg-white text-slate-900 font-bold rounded-full hover:scale-105 hover:bg-slate-50 transition-all flex items-center gap-2 shadow-xl"
                >
                   Next <ArrowRight className="w-5 h-5" />
                </button>
              </div>
            )}

            {/* Resign Confirmation Overlay */}
            {showResignConfirm && !winner && (
              <div className="absolute inset-0 z-50 bg-black/70 flex flex-col items-center justify-center backdrop-blur-md animate-in fade-in p-6 text-center">
                <div className="bg-white dark:bg-slate-800 p-8 rounded-2xl shadow-2xl max-w-sm w-full mx-4 border border-slate-200 dark:border-slate-700">
                  <div className="w-16 h-16 bg-red-100 dark:bg-red-900/30 rounded-full flex items-center justify-center mx-auto mb-4">
                     <AlertTriangle className="w-8 h-8 text-red-600 dark:text-red-500" />
                  </div>
                  <h3 className="text-2xl font-bold text-slate-900 dark:text-white mb-2">Resign Game?</h3>
                  <p className="text-slate-500 dark:text-slate-400 mb-8">
                    To start a new game or join another, you must resign the current match.
                  </p>
                  <div className="flex gap-3 justify-center">
                      <button 
                          onClick={() => setShowResignConfirm(false)}
                          className="px-6 py-2.5 rounded-full bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-200 font-bold hover:bg-slate-200 dark:hover:bg-slate-600 transition-colors w-full"
                      >
                          Cancel
                      </button>
                      <button 
                          onClick={confirmResignation}
                          className="px-6 py-2.5 rounded-full bg-red-600 text-white font-bold hover:bg-red-700 transition-colors flex items-center justify-center gap-2 w-full"
                      >
                          Resign
                      </button>
                  </div>
                </div>
              </div>
            )}

            {/* Game Menu (Create/Join/Local) Overlay */}
            {showGameMenu && (
                <div className="absolute inset-0 z-50 bg-slate-900/90 flex flex-col items-center justify-center backdrop-blur-md animate-in fade-in p-6 text-center">
                    <div className="bg-white dark:bg-slate-800 p-6 md:p-8 rounded-3xl shadow-2xl max-w-md w-full border border-slate-200 dark:border-slate-700">
                        <h2 className="text-3xl font-bold text-slate-900 dark:text-white mb-2">New Game</h2>
                        <p className="text-slate-500 dark:text-slate-400 mb-8">Choose how you want to play</p>
                        
                        <div className="space-y-3">
                            {/* Option 1: Play Local */}
                            <button 
                                onClick={resetLocalGame}
                                className="w-full flex items-center p-4 rounded-xl border-2 border-slate-100 dark:border-slate-700 hover:border-blue-500 dark:hover:border-blue-500 hover:bg-blue-50 dark:hover:bg-slate-700 transition-all group"
                            >
                                <div className="p-3 bg-blue-100 dark:bg-blue-900/50 rounded-full text-blue-600 dark:text-blue-400 mr-4 group-hover:scale-110 transition-transform">
                                    <Users className="w-6 h-6" />
                                </div>
                                <div className="text-left">
                                    <h3 className="font-bold text-slate-900 dark:text-white">Play Local</h3>
                                    <p className="text-xs text-slate-500 dark:text-slate-400">Pass and play on this device</p>
                                </div>
                            </button>

                            {/* Option 2: Create Online */}
                            <button 
                                onClick={createOnlineGame}
                                className="w-full flex items-center p-4 rounded-xl border-2 border-slate-100 dark:border-slate-700 hover:border-green-500 dark:hover:border-green-500 hover:bg-green-50 dark:hover:bg-slate-700 transition-all group"
                            >
                                <div className="p-3 bg-green-100 dark:bg-green-900/50 rounded-full text-green-600 dark:text-green-400 mr-4 group-hover:scale-110 transition-transform">
                                    <Globe className="w-6 h-6" />
                                </div>
                                <div className="text-left">
                                    <h3 className="font-bold text-slate-900 dark:text-white">Create Online Game</h3>
                                    <p className="text-xs text-slate-500 dark:text-slate-400">Get a Game ID to share with a friend</p>
                                </div>
                            </button>

                            {/* Option 3: Join Online */}
                            <div className="p-4 rounded-xl border-2 border-slate-100 dark:border-slate-700 bg-slate-50 dark:bg-slate-900/50">
                                <h3 className="font-bold text-slate-900 dark:text-white text-left mb-2 flex items-center gap-2">
                                    <Play className="w-4 h-4" /> Join Game
                                </h3>
                                <div className="flex gap-2">
                                    <input 
                                        type="text" 
                                        placeholder="Enter Game ID" 
                                        value={joinId}
                                        onChange={(e) => setJoinId(e.target.value.toUpperCase())}
                                        className="flex-1 bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-600 rounded-lg px-3 py-2 text-sm font-mono uppercase focus:ring-2 focus:ring-blue-500 outline-none dark:text-white"
                                    />
                                    <button 
                                        onClick={joinOnlineGame}
                                        disabled={!joinId}
                                        className="bg-blue-600 text-white px-4 py-2 rounded-lg font-bold text-sm hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                                    >
                                        Join
                                    </button>
                                </div>
                            </div>
                        </div>

                         <button 
                            onClick={() => setShowGameMenu(false)}
                            className="mt-6 text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-200 text-sm font-medium"
                        >
                            Cancel
                        </button>
                    </div>
                </div>
            )}

            <div className="w-full h-full grid grid-cols-8 grid-rows-8">
              {board.map((row, r) => (
                row.map((piece, c) => {
                  const isValid = validMoves.some(m => m.r === r && m.c === c);
                  const isSelected = selected?.r === r && selected?.c === c;
                  const isCheckSquare = check && piece && getPieceColor(piece) === turn && piece.toLowerCase() === 'k';

                  return (
                    <div
                      key={`${r}-${c}`}
                      onClick={() => handleSquareClick(r, c)}
                      style={{ backgroundColor: getSquareColor(r, c) }}
                      className="relative flex items-center justify-center cursor-pointer"
                    >
                      {/* Rank/File Indicators */}
                      {c === 0 && r === 7 && <span className={`absolute bottom-0.5 left-1 text-[10px] font-bold ${getPieceColor(piece) === 'white' ? 'text-slate-800' : 'opacity-60'}`}>a1</span>}
                      
                      {isSelected && <div className="absolute inset-0 bg-yellow-400/40" />}
                      {isValid && !piece && <div className="w-4 h-4 rounded-full bg-black/20 dark:bg-black/40" />}
                      {isValid && piece && (
                         <div className="absolute inset-0">
                           <div className="absolute top-0 left-0 w-3 h-3 border-t-4 border-l-4 border-black/20" />
                           <div className="absolute top-0 right-0 w-3 h-3 border-t-4 border-r-4 border-black/20" />
                           <div className="absolute bottom-0 left-0 w-3 h-3 border-b-4 border-l-4 border-black/20" />
                           <div className="absolute bottom-0 right-0 w-3 h-3 border-b-4 border-r-4 border-black/20" />
                         </div>
                      )}
                      {isCheckSquare && (
                        <div className="absolute inset-0" style={{ background: 'radial-gradient(circle, rgba(255,0,0,0.8) 0%, rgba(255,0,0,0) 70%)' }} />
                      )}
                      {piece && (
                        <div className="w-4/5 h-4/5 z-10">
                          <PieceComponent type={piece} color={getPieceColor(piece)} />
                        </div>
                      )}
                    </div>
                  );
                })
              ))}
            </div>
          </div>
          
          <div className="bg-slate-200 dark:bg-slate-800 text-slate-700 dark:text-slate-200 p-4 rounded-xl flex flex-col sm:flex-row justify-between items-center shadow-inner gap-3 border-2 border-slate-300 dark:border-slate-600">
             <div className="flex items-center gap-3">
               <div className={`w-3 h-3 rounded-full ${turn === 'white' ? 'bg-white border border-slate-300' : 'bg-slate-900 border border-slate-600'}`}></div>
               <span className="font-semibold capitalize">{turn}'s Turn</span>
               {check && !winner && <span className="text-red-500 font-bold ml-2 animate-pulse">CHECK!</span>}
             </div>
             
             {/* Game ID Display (Only when online) */}
             {gameId && (
                 <div className="flex items-center gap-2 bg-white dark:bg-slate-700 px-3 py-1.5 rounded-lg border border-slate-300 dark:border-slate-600">
                     <span className="text-xs font-bold text-slate-500 dark:text-slate-300 uppercase">ID:</span>
                     <span className="font-mono font-bold text-blue-600 dark:text-blue-400 tracking-wider">{gameId}</span>
                     <button onClick={copyGameId} className="hover:text-blue-500 p-1"><Copy className="w-3 h-3" /></button>
                 </div>
             )}

             <button 
                onClick={onNewGameClick} 
                className="text-sm font-medium hover:text-red-500 flex items-center gap-1 transition-colors"
             >
               <Flag className="w-4 h-4" /> New Game
             </button>
          </div>
        </div>

        {/* --- RIGHT PANEL --- */}
        {showSettings && (
          <div className="w-full md:w-80 flex flex-col gap-6 animate-in slide-in-from-right-10 fade-in duration-300">
            <div className="flex items-center justify-between mb-2">
              <h2 className="text-xl font-bold tracking-tight text-slate-900 dark:text-slate-100">Settings</h2>
              <button 
                onClick={() => setShowSettings(false)}
                className="w-10 h-10 flex items-center justify-center rounded-full bg-slate-200 dark:bg-slate-800 text-slate-700 dark:text-slate-200 hover:bg-red-100 hover:text-red-500 dark:hover:bg-red-900/30 dark:hover:text-red-400 transition-colors"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="bg-white/95 dark:bg-slate-800/95 text-slate-800 dark:text-slate-200 p-6 rounded-2xl shadow-sm backdrop-blur-md border-2 border-slate-300 dark:border-slate-600 space-y-6">
              <div className="flex items-center justify-between">
                <label className="text-xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400 flex items-center gap-2">App Theme</label>
                <button onClick={() => setDarkMode(!darkMode)} className={`p-2 rounded-full transition-colors ${darkMode ? 'bg-slate-800 text-yellow-400' : 'bg-slate-200 text-slate-700'}`}>
                    {darkMode ? <Sun className="w-5 h-5" /> : <Moon className="w-5 h-5" />}
                </button>
              </div>
              <hr className="border-slate-300 dark:border-slate-700" />
              <div className="space-y-3">
                <label className="text-xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400 flex items-center gap-2"><Palette className="w-4 h-4" /> Board Theme</label>
                <div className="grid grid-cols-2 gap-2">
                  {Object.entries(BOARD_THEMES).map(([key, theme]) => (
                    <button key={key} onClick={() => setBoardTheme(key)} className={`relative overflow-hidden h-12 rounded-lg border-2 transition-all flex items-center justify-center ${boardTheme === key ? 'border-blue-500 scale-105 shadow-md' : 'border-slate-300 dark:border-slate-600 hover:border-slate-400 opacity-70 hover:opacity-100'}`}>
                      <div className="absolute inset-0 flex">
                        <div className="w-1/2 h-full" style={{ backgroundColor: theme.light }} />
                        <div className="w-1/2 h-full" style={{ backgroundColor: theme.dark }} />
                      </div>
                      <span className={`relative z-10 text-[10px] font-bold px-2 py-1 rounded bg-black/40 text-white backdrop-blur-sm`}>{theme.name}</span>
                    </button>
                  ))}
                </div>
              </div>
              <hr className="border-slate-300 dark:border-slate-700" />
              <div className="space-y-3">
                <label className="text-xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400 flex items-center gap-2"><LayoutGrid className="w-4 h-4" /> Piece Style</label>
                <div className="flex flex-col gap-2">
                  {Object.keys(Pieces).map((key) => (
                    <button key={key} onClick={() => setPieceTheme(key)} className={`flex items-center gap-3 p-2 rounded-lg border-2 transition-all text-slate-700 dark:text-slate-200 ${pieceTheme === key ? 'border-blue-500 bg-blue-50 dark:bg-slate-700' : 'border-slate-300 dark:border-slate-600 hover:bg-slate-100 dark:hover:bg-slate-700 hover:border-slate-400'}`}>
                      <div className="w-8 h-8 text-slate-800 dark:text-slate-200">
                      {React.createElement(Pieces[key].N, {
                          stroke: key === 'Neo' ? 'currentColor' : undefined,
                          fill: key === 'Neo' ? 'none' : 'currentColor' 
                        })}
                      </div>
                      <span className="font-medium text-sm">{key}</span>
                    </button>
                  ))}
                </div>
              </div>
            </div>
            <div className="text-xs text-center text-slate-400 dark:text-slate-500 mt-auto">Lite version. En Passant & Castling coming soon.</div>
          </div>
        )}
      </div>
    </div>
  );
};

export default App;