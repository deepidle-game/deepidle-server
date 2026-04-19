const GAME_API_URL = "http://localhost:3000/api";
let authToken = "";

async function fetchGameInfo(endpoint, method = "GET", body = null) {
  const headers = { "Content-Type": "application/json" };
  if (authToken) headers["Authorization"] = `Bearer ${authToken}`;

  const res = await fetch(`${GAME_API_URL}${endpoint}`, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  });

  if (!res.ok) {
    const errorText = await res.text();
    throw new Error(`${endpoint} Error: ${res.status} - ${errorText}`);
  }
  return res.json();
}

async function play() {
  const username = "ai_player_" + Math.floor(Math.random() * 1000);
  const password = "password123";

  console.log(`\n🤖 [1] Creating character: ${username}`);
  await fetchGameInfo("/auth/signup", "POST", { username, password });
  const loginRes = await fetchGameInfo("/auth/signin", "POST", { username, password });
  authToken = loginRes.token;

  console.log(`🤖 [2] Checking character status...`);
  const status = await fetchGameInfo("/character/detail", "GET");
  console.log("Status:", status);

  console.log(`\n🤖 [3] Checking inventory...`);
  const inv = await fetchGameInfo("/inventory", "GET");
  console.log("Inventory:", inv);

  console.log(`\n🤖 [4] Setting action to 'cutting_wood'...`);
  await fetchGameInfo("/character/action", "POST", { action: "cutting_wood" });

  console.log(`🤖 [5] Waiting for 3 seconds to gather resources...`);
  await new Promise(r => setTimeout(r, 3000));

  console.log(`\n🤖 [6] Claiming resources...`);
  const claimRes = await fetchGameInfo("/character/claim", "POST");
  console.log("Claim Result:", claimRes);

  console.log(`\n🤖 [7] Checking inventory after claim...`);
  const invAfter = await fetchGameInfo("/inventory", "GET");
  console.log("Updated Inventory:", invAfter);
  
  console.log(`\n🤖 [8] Upgrading wooden_sword using the gathered resources...`);
  try {
      const upgradeRes = await fetchGameInfo("/inventory/upgrade", "POST", { item_id: "wooden_sword" });
      console.log("Upgrade Result:", upgradeRes);
  } catch(e) {
      console.log("Upgrade Failed (need more resources or other reason):", e.message);
  }

  const finalInv = await fetchGameInfo("/inventory", "GET");
  console.log("Final Inventory:", finalInv);
}

play().catch(console.error);
