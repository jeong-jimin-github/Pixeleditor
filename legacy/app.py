import tkinter as tk
from tkinter import filedialog, colorchooser, simpledialog, messagebox, Scale, Spinbox
from PIL import Image, ImageTk, ImageDraw

class ProfessionalPixelEditor:
    def __init__(self, root):
        self.root = root
        self.root.title("Pixel Editor Pro - Palette & Save")
        self.root.geometry("1200x850")

        # --- 設定 ---
        self.image_width = 64
        self.image_height = 64
        self.zoom = 10
        self.min_zoom = 1
        self.max_zoom = 40
        
        self.brush_color = (0, 0, 0, 255)
        self.secondary_color = (0, 0, 0, 0)
        self.current_tool = "pen"
        self.brush_size = 1
        
        self.show_grid = True
        self.grid_gap = 1
        self.symmetry_mode = False

        # ファイル管理用
        self.file_path = None  # 現在開いている/保存したパス

        # 履歴
        self.history = []
        self.redo_stack = []
        self.max_history = 30

        # メイン画像
        self.image = Image.new("RGBA", (self.image_width, self.image_height), (0, 0, 0, 0))
        self.draw = ImageDraw.Draw(self.image)

        # 座標管理
        self.start_x = None
        self.start_y = None
        self.temp_items = [] 
        self.last_cursor_x = -1
        self.last_cursor_y = -1
        self.is_panning = False
        self.pan_start_x = 0
        self.pan_start_y = 0

        self.create_ui()
        self.setup_shortcuts()
        
        self.save_state()
        self.create_checkerboard_pattern()
        self.update_canvas()
        self.refresh_palette() # 初期パレット生成

    def create_ui(self):
        main_frame = tk.Frame(self.root)
        main_frame.pack(fill=tk.BOTH, expand=True)

        # --- サイドバー ---
        tools_frame = tk.Frame(main_frame, width=140, bg="#e0e0e0", relief=tk.RAISED, bd=1)
        tools_frame.pack(side=tk.LEFT, fill=tk.Y)

        # ツール群
        tk.Label(tools_frame, text="Tools", font=("Arial", 10, "bold"), bg="#e0e0e0").pack(pady=5)
        self.create_tool_btn(tools_frame, "Pen (P)", "pen")
        self.create_tool_btn(tools_frame, "Eraser (E)", "eraser")
        self.create_tool_btn(tools_frame, "Fill (F)", "fill")
        self.create_tool_btn(tools_frame, "Pipette (I)", "pipette")

        # 対称モード
        self.symmetry_var = tk.BooleanVar(value=False)
        tk.Checkbutton(tools_frame, text="Symmetry X", variable=self.symmetry_var, 
                       command=self.toggle_symmetry, bg="#e0e0e0").pack(pady=(5, 0))

        # カラー選択
        tk.Label(tools_frame, text="Color", bg="#e0e0e0").pack(pady=(10,0))
        self.color_btn = tk.Button(tools_frame, text="     ", bg="black", command=self.choose_color, relief=tk.SUNKEN, bd=3)
        self.color_btn.pack(pady=5, ipadx=10, ipady=5)

        # --- パレットエリア (NEW) ---
        tk.Label(tools_frame, text="Palette", font=("Arial", 10, "bold"), bg="#e0e0e0").pack(pady=(15, 0))
        
        # パレット操作ボタン
        palette_ctrl_frame = tk.Frame(tools_frame, bg="#e0e0e0")
        palette_ctrl_frame.pack(fill=tk.X, padx=5)
        tk.Button(palette_ctrl_frame, text="Refresh", command=self.refresh_palette, font=("Arial", 8)).pack(fill=tk.X, pady=2)
        
        # パレット制限モード
        self.lock_palette_var = tk.BooleanVar(value=False)
        tk.Checkbutton(tools_frame, text="Lock Palette", variable=self.lock_palette_var, bg="#e0e0e0", font=("Arial", 8)).pack()

        # パレット表示用フレーム（スクロール付き）
        self.palette_container = tk.Frame(tools_frame, bg="#d0d0d0", relief=tk.SUNKEN, bd=1)
        self.palette_container.pack(fill=tk.BOTH, expand=True, padx=5, pady=5)

        # --- 上部バー ---
        top_bar = tk.Frame(main_frame, bg="#d0d0d0", height=40)
        top_bar.pack(side=tk.TOP, fill=tk.X)

        self.create_top_btn(top_bar, "Open", self.load_image)
        self.create_top_btn(top_bar, "Save (Ctrl+S)", self.quick_save) # 上書き保存
        self.create_top_btn(top_bar, "Save As...", self.save_image_as) # 名前をつけて保存
        
        tk.Frame(top_bar, width=15, bg="#d0d0d0").pack(side=tk.LEFT)
        self.create_top_btn(top_bar, "Undo", self.undo)
        self.create_top_btn(top_bar, "Redo", self.redo)
        self.create_top_btn(top_bar, "Clear", self.clear_canvas)

        tk.Frame(top_bar, width=15, bg="#d0d0d0").pack(side=tk.LEFT)
        self.create_top_btn(top_bar, "Grid:", self.toggle_grid)
        self.grid_spin = Spinbox(top_bar, from_=1, to=128, width=3, command=self.change_grid_gap)
        self.grid_spin.pack(side=tk.LEFT, padx=2)
        
        tk.Frame(top_bar, width=15, bg="#d0d0d0").pack(side=tk.LEFT)
        tk.Button(top_bar, text="-", command=lambda: self.change_zoom(-1), width=2).pack(side=tk.LEFT)
        self.zoom_label = tk.Label(top_bar, text=f"{self.zoom}x", bg="#d0d0d0", width=4)
        self.zoom_label.pack(side=tk.LEFT)
        tk.Button(top_bar, text="+", command=lambda: self.change_zoom(1), width=2).pack(side=tk.LEFT)

        # --- キャンバスエリア ---
        self.canvas_area = tk.Frame(main_frame, bg="#303030")
        self.canvas_area.pack(fill=tk.BOTH, expand=True)

        self.v_scroll = tk.Scrollbar(self.canvas_area, orient=tk.VERTICAL)
        self.h_scroll = tk.Scrollbar(self.canvas_area, orient=tk.HORIZONTAL)

        self.canvas = tk.Canvas(self.canvas_area, bg="#303030", highlightthickness=0,
                                xscrollcommand=self.h_scroll.set, yscrollcommand=self.v_scroll.set)
        
        self.v_scroll.config(command=self.canvas.yview)
        self.h_scroll.config(command=self.canvas.xview)

        self.v_scroll.pack(side=tk.RIGHT, fill=tk.Y)
        self.h_scroll.pack(side=tk.BOTTOM, fill=tk.X)
        self.canvas.pack(side=tk.LEFT, fill=tk.BOTH, expand=True)

        # イベントバインド
        self.canvas.bind("<Button-1>", self.on_click)
        self.canvas.bind("<B1-Motion>", self.on_drag)
        self.canvas.bind("<ButtonRelease-1>", self.on_release)
        self.canvas.bind("<Button-3>", self.pick_color_from_canvas)
        self.canvas.bind("<Control-MouseWheel>", self.on_mouse_wheel)
        self.canvas.bind("<Motion>", self.on_mouse_move)
        
        self.root.bind("<space>", self.start_pan_mode)
        self.root.bind("<KeyRelease-space>", self.end_pan_mode)

    def create_tool_btn(self, parent, text, tool):
        tk.Button(parent, text=text, command=lambda: self.set_tool(tool), width=12).pack(pady=2)

    def create_top_btn(self, parent, text, cmd):
        tk.Button(parent, text=text, command=cmd).pack(side=tk.LEFT, padx=2, pady=3)

    # --- パレット機能 (NEW) ---

    def refresh_palette(self):
        """画像から色を抽出してパレットUIを更新"""
        # 既存のパレットボタンを削除
        for widget in self.palette_container.winfo_children():
            widget.destroy()

        # 画像内の色を取得 (maxcolorsを大きめに設定して全色取得)
        # getcolorsは (count, pixel) のリストを返す
        colors = self.image.getcolors(maxcolors=100000)
        
        if not colors:
            return

        # 出現頻度順にソート (オプション)
        colors = sorted(colors, key=lambda x: x[0], reverse=True)

        # グリッド配置用変数
        row = 0
        col = 0
        max_cols = 4 # 1行に並べる数

        seen_colors = set()

        for count, rgba in colors:
            # 透明色(Alpha=0)は無視、または特別扱い
            if rgba[3] == 0:
                continue
            
            # 重複回避（念のため）
            if rgba in seen_colors:
                continue
            seen_colors.add(rgba)

            # 色変換
            hex_color = '#{:02x}{:02x}{:02x}'.format(rgba[0], rgba[1], rgba[2])
            
            # パレットボタン作成
            btn = tk.Button(self.palette_container, bg=hex_color, width=2, height=1,
                            command=lambda c=rgba, h=hex_color: self.set_palette_color(c, h))
            btn.grid(row=row, column=col, padx=1, pady=1)
            
            # ツールチップ的に色情報を表示（簡易）
            # btn.bind("<Enter>", lambda e, h=hex_color: print(h))

            col += 1
            if col >= max_cols:
                col = 0
                row += 1
                
    def set_palette_color(self, rgba, hex_color):
        """パレットボタンから色を選択"""
        self.brush_color = rgba
        self.color_btn.config(bg=hex_color)
        self.set_tool("pen")

    def choose_color(self):
        """カラーピッカーを開く（制限モードなら拒否）"""
        if self.lock_palette_var.get():
            messagebox.showwarning("Locked", "Palette is locked.\nPlease select a color from the Palette list.")
            return

        color = colorchooser.askcolor()
        if color[1]:
            r, g, b = [int(c) for c in color[0]]
            self.brush_color = (r, g, b, 255)
            self.color_btn.config(bg=color[1])
            # 色を変えたらパレットも更新する？（好みによるが、今回はRefreshボタン手動更新とする）

    # --- 保存・読み込み (Updated) ---

    def quick_save(self, event=None):
        """Ctrl+S: 上書き保存、パスがなければ名前を付けて保存"""
        if self.file_path:
            try:
                self.image.save(self.file_path)
                messagebox.showinfo("Saved", f"Saved to {self.file_path}")
            except Exception as e:
                messagebox.showerror("Error", str(e))
        else:
            self.save_image_as()

    def save_image_as(self):
        """名前を付けて保存"""
        path = filedialog.asksaveasfilename(defaultextension=".png", filetypes=[("PNG", "*.png")])
        if path:
            self.file_path = path
            try:
                self.image.save(path)
                self.root.title(f"Pixel Editor Pro - {path}")
                messagebox.showinfo("Saved", "Image saved.")
            except Exception as e:
                messagebox.showerror("Error", str(e))

    def load_image(self):
        path = filedialog.askopenfilename(filetypes=[("Images", "*.png *.jpg *.jpeg *.bmp")])
        if path:
            self.save_state()
            try:
                img = Image.open(path).convert("RGBA")
                self.image = img
                self.image_width, self.image_height = img.size
                
                # ファイルパスを記録
                self.file_path = path
                self.root.title(f"Pixel Editor Pro - {path}")
                
                if hasattr(self, 'bg_pattern_img'):
                    del self.bg_pattern_img
                
                self.draw = ImageDraw.Draw(self.image)
                if self.image_width > 100:
                    self.zoom = 4
                
                self.update_canvas()
                
                # 画像読み込み時に自動でパレット生成
                self.refresh_palette()
                
            except Exception as e:
                messagebox.showerror("Error", str(e))

    # --- 既存の描画・操作ロジック ---

    def create_checkerboard_pattern(self):
        self.bg_pattern_img = Image.new("RGB", (self.image_width, self.image_height), (255,255,255))
        pixels = self.bg_pattern_img.load()
        for y in range(self.image_height):
            for x in range(self.image_width):
                if (x + y) % 2 == 1:
                    pixels[x, y] = (204, 204, 204)

    def update_canvas(self):
        display_w = self.image_width * self.zoom
        display_h = self.image_height * self.zoom
        
        if not hasattr(self, 'bg_pattern_img') or self.bg_pattern_img.size != (self.image_width, self.image_height):
            self.create_checkerboard_pattern()
            
        resized_bg = self.bg_pattern_img.resize((display_w, display_h), Image.NEAREST)
        self.tk_bg_pattern = ImageTk.PhotoImage(resized_bg)
        
        resized_img = self.image.resize((display_w, display_h), Image.NEAREST)
        self.tk_image = ImageTk.PhotoImage(resized_img)
        
        self.canvas.delete("all")
        self.canvas.create_image(0, 0, anchor=tk.NW, image=self.tk_bg_pattern, tags="bg")
        self.canvas.create_image(0, 0, anchor=tk.NW, image=self.tk_image, tags="img")
        
        if self.show_grid:
            grid_step = self.grid_gap * self.zoom
            if grid_step >= 4:
                for x in range(0, display_w + 1, grid_step):
                    self.canvas.create_line(x, 0, x, display_h, fill="#888888", tags="grid")
                for y in range(0, display_h + 1, grid_step):
                    self.canvas.create_line(0, y, display_w, y, fill="#888888", tags="grid")
            if self.symmetry_mode:
                center_x = (self.image_width // 2) * self.zoom
                self.canvas.create_line(center_x, 0, center_x, display_h, fill="blue", width=2, tags="grid")

        self.canvas.config(scrollregion=(0, 0, display_w, display_h))
        self.zoom_label.config(text=f"{self.zoom}x")

    def on_mouse_move(self, event):
        if self.is_panning: return
        x, y = self.get_pixel_coords(event)
        if x == self.last_cursor_x and y == self.last_cursor_y:
            return
        self.last_cursor_x, self.last_cursor_y = x, y
        self.update_cursor_highlight(x, y)

    def update_cursor_highlight(self, x, y):
        self.canvas.delete("cursor")
        if x < 0 or x >= self.image_width or y < 0 or y >= self.image_height:
            return
        sx = x * self.zoom
        sy = y * self.zoom
        ex = sx + (self.zoom * self.brush_size)
        ey = sy + (self.zoom * self.brush_size)
        offset = (self.brush_size // 2) * self.zoom
        self.canvas.create_rectangle(sx-offset, sy-offset, ex-offset, ey-offset, outline="red", width=2, tags="cursor")
        if self.symmetry_mode:
            sym_x = self.image_width - 1 - x
            sx_sym = sym_x * self.zoom - offset
            self.canvas.create_rectangle(sx_sym, sy-offset, sx_sym + (self.zoom * self.brush_size), ey-offset, outline="blue", width=2, dash=(2, 2), tags="cursor")

    def change_zoom(self, delta):
        new_zoom = self.zoom + delta
        if self.min_zoom <= new_zoom <= self.max_zoom:
            self.zoom = new_zoom
            self.update_canvas()

    def on_mouse_wheel(self, event):
        if event.delta > 0: self.change_zoom(1)
        else: self.change_zoom(-1)

    def get_pixel_coords(self, event):
        canvas_x = self.canvas.canvasx(event.x)
        canvas_y = self.canvas.canvasy(event.y)
        x = int(canvas_x // self.zoom)
        y = int(canvas_y // self.zoom)
        return x, y

    def on_click(self, event):
        if self.is_panning:
            self.pan_start_x = event.x
            self.pan_start_y = event.y
            return
        self.save_state()
        x, y = self.get_pixel_coords(event)
        if not (0 <= x < self.image_width and 0 <= y < self.image_height): return
        self.start_x, self.start_y = x, y
        if self.current_tool == "pen": self.draw_brush(x, y)
        elif self.current_tool == "eraser": self.draw_brush(x, y, erase=True)
        elif self.current_tool == "fill": self.flood_fill(x, y)
        elif self.current_tool == "pipette": self.pick_color(x, y)

    def on_drag(self, event):
        if self.is_panning:
            dx = self.pan_start_x - event.x
            dy = self.pan_start_y - event.y
            self.canvas.xview_scroll(dx, "units")
            self.canvas.yview_scroll(dy, "units")
            self.pan_start_x, self.pan_start_y = event.x, event.y
            return
        x, y = self.get_pixel_coords(event)
        self.update_cursor_highlight(x, y)
        if not (0 <= x < self.image_width and 0 <= y < self.image_height): return
        if self.current_tool == "pen": self.draw_brush(x, y)
        elif self.current_tool == "eraser": self.draw_brush(x, y, erase=True)

    def on_release(self, event):
        if self.is_panning: return
        self.start_x = None
        
        # 描画が終わったタイミングでパレットを更新するか？
        # 自動更新すると重くなる可能性があるため、今回は手動(Refresh)またはロード時のみに限定
        # self.refresh_palette() 

    def draw_brush(self, x, y, erase=False):
        color = self.secondary_color if erase else self.brush_color
        self.draw_point_internal(x, y, color)
        if self.symmetry_mode:
            sym_x = self.image_width - 1 - x
            if sym_x != x: self.draw_point_internal(sym_x, y, color)
        self.update_canvas()

    def draw_point_internal(self, x, y, color):
        if self.brush_size == 1: self.draw.point((x, y), fill=color)
        else:
            r = self.brush_size // 2
            self.draw.rectangle([x-r, y-r, x+r, y+r], fill=color)

    def flood_fill(self, x, y):
        try:
            ImageDraw.floodfill(self.image, (x, y), self.brush_color, thresh=0)
            if self.symmetry_mode:
                ImageDraw.floodfill(self.image, (self.image_width - 1 - x, y), self.brush_color, thresh=0)
            self.update_canvas()
            # 塗りつぶしで新色が増える可能性があるので、必要ならここで refresh_palette()
        except: pass

    def pick_color(self, x, y):
        r, g, b, a = self.image.getpixel((x, y))
        # パレットロック中であっても、スポイト機能（キャンバス上の色）は許可するのが一般的ですが、
        # もしスポイトも禁止したい場合はここにチェックを入れる
        self.brush_color = (r, g, b, 255)
        hex_color = '#{:02x}{:02x}{:02x}'.format(r, g, b)
        self.color_btn.config(bg=hex_color)
        self.set_tool("pen")

    def pick_color_from_canvas(self, event):
        x, y = self.get_pixel_coords(event)
        self.pick_color(x, y)

    def set_tool(self, tool): self.current_tool = tool
    def change_brush_size(self, val): self.brush_size = int(val)
    def change_grid_gap(self):
        try:
            self.grid_gap = max(1, int(self.grid_spin.get()))
            self.update_canvas()
        except: pass
    def toggle_grid(self):
        self.show_grid = not self.show_grid
        self.update_canvas()
    def toggle_symmetry(self):
        self.symmetry_mode = self.symmetry_var.get()
        self.update_canvas()
    def start_pan_mode(self, event):
        self.is_panning = True
        self.canvas.config(cursor="fleur")
    def end_pan_mode(self, event):
        self.is_panning = False
        self.canvas.config(cursor="")
    def clear_canvas(self):
        self.save_state()
        self.image = Image.new("RGBA", (self.image_width, self.image_height), (0, 0, 0, 0))
        self.draw = ImageDraw.Draw(self.image)
        self.update_canvas()
        self.refresh_palette()
    def save_state(self):
        self.history.append(self.image.copy())
        if len(self.history) > self.max_history: self.history.pop(0)
        self.redo_stack.clear()
    def undo(self):
        if self.history:
            self.redo_stack.append(self.image.copy())
            self.image = self.history.pop()
            self.draw = ImageDraw.Draw(self.image)
            self.update_canvas()
            self.refresh_palette()
    def redo(self):
        if self.redo_stack:
            self.history.append(self.image.copy())
            self.image = self.redo_stack.pop()
            self.draw = ImageDraw.Draw(self.image)
            self.update_canvas()
            self.refresh_palette()
    def setup_shortcuts(self):
        self.root.bind("<Control-z>", lambda e: self.undo())
        self.root.bind("<Control-y>", lambda e: self.redo())
        self.root.bind("<Control-s>", self.quick_save) # ショートカット登録
        self.root.bind("p", lambda e: self.set_tool("pen"))
        self.root.bind("e", lambda e: self.set_tool("eraser"))
        self.root.bind("f", lambda e: self.set_tool("fill"))
        self.root.bind("i", lambda e: self.set_tool("pipette"))

if __name__ == "__main__":
    root = tk.Tk()
    app = ProfessionalPixelEditor(root)
    root.mainloop()