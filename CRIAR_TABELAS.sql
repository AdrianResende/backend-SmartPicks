-- =====================================================
-- SCRIPT DE CRIAÇÃO DAS TABELAS - SmartPicks
-- Execute este script no pgAdmin ou qualquer cliente PostgreSQL
-- Banco de dados: smartpicks
-- =====================================================

-- =====================================================
-- PASSO 1: Criar a função de update_updated_at (se não existir)
-- =====================================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- =====================================================
-- PASSO 2: Criar tabela PALPITES
-- =====================================================

CREATE TABLE IF NOT EXISTS palpites (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    titulo VARCHAR(255),
    img_url TEXT,
    link TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign key
    CONSTRAINT fk_palpite_user 
        FOREIGN KEY (user_id) 
        REFERENCES users(id) 
        ON DELETE CASCADE
);

-- Índices para palpites
CREATE INDEX IF NOT EXISTS idx_palpites_user_id ON palpites (user_id);
CREATE INDEX IF NOT EXISTS idx_palpites_created_at ON palpites (created_at DESC);

-- Trigger para updated_at em palpites
DROP TRIGGER IF EXISTS update_palpites_updated_at ON palpites;
CREATE TRIGGER update_palpites_updated_at
    BEFORE UPDATE ON palpites
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- =====================================================
-- PASSO 3: Criar tabela COMENTÁRIOS
-- =====================================================

CREATE TABLE IF NOT EXISTS comentarios (
    id SERIAL PRIMARY KEY,
    palpite_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    texto TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign keys
    CONSTRAINT fk_comentario_palpite 
        FOREIGN KEY (palpite_id) 
        REFERENCES palpites(id) 
        ON DELETE CASCADE,
    
    CONSTRAINT fk_comentario_user 
        FOREIGN KEY (user_id) 
        REFERENCES users(id) 
        ON DELETE CASCADE
);

-- Índices para comentários
CREATE INDEX IF NOT EXISTS idx_comentarios_palpite_id ON comentarios (palpite_id);
CREATE INDEX IF NOT EXISTS idx_comentarios_user_id ON comentarios (user_id);
CREATE INDEX IF NOT EXISTS idx_comentarios_created_at ON comentarios (created_at DESC);

-- Trigger para updated_at em comentários
DROP TRIGGER IF EXISTS update_comentarios_updated_at ON comentarios;
CREATE TRIGGER update_comentarios_updated_at
    BEFORE UPDATE ON comentarios
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- =====================================================
-- PASSO 4: Criar tabela PALPITES_REACTIONS
-- =====================================================

CREATE TABLE IF NOT EXISTS palpites_reactions (
    id SERIAL PRIMARY KEY,
    palpite_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    tipo VARCHAR(10) NOT NULL CHECK (tipo IN ('like', 'dislike')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign keys
    CONSTRAINT fk_reaction_palpite 
        FOREIGN KEY (palpite_id) 
        REFERENCES palpites(id) 
        ON DELETE CASCADE,
    
    CONSTRAINT fk_reaction_user 
        FOREIGN KEY (user_id) 
        REFERENCES users(id) 
        ON DELETE CASCADE,
    
    -- Um usuário pode ter apenas uma reação por palpite
    CONSTRAINT unique_user_palpite_reaction 
        UNIQUE (palpite_id, user_id)
);

-- Índices para reactions
CREATE INDEX IF NOT EXISTS idx_reactions_palpite_id ON palpites_reactions (palpite_id);
CREATE INDEX IF NOT EXISTS idx_reactions_user_id ON palpites_reactions (user_id);
CREATE INDEX IF NOT EXISTS idx_reactions_tipo ON palpites_reactions (tipo);

-- =====================================================
-- PASSO 5: Criar tabela COMENTARIOS_REACTIONS
-- =====================================================

CREATE TABLE IF NOT EXISTS comentarios_reactions (
    id SERIAL PRIMARY KEY,
    comentario_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    tipo VARCHAR(10) NOT NULL CHECK (tipo IN ('like', 'dislike')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign keys
    CONSTRAINT fk_comentario_reaction_comentario 
        FOREIGN KEY (comentario_id) 
        REFERENCES comentarios(id) 
        ON DELETE CASCADE,
    
    CONSTRAINT fk_comentario_reaction_user 
        FOREIGN KEY (user_id) 
        REFERENCES users(id) 
        ON DELETE CASCADE,
    
    -- Um usuário pode ter apenas uma reação por comentário
    CONSTRAINT unique_user_comentario_reaction 
        UNIQUE (comentario_id, user_id)
);

-- Índices para comentarios_reactions
CREATE INDEX IF NOT EXISTS idx_comentarios_reactions_comentario_id ON comentarios_reactions (comentario_id);
CREATE INDEX IF NOT EXISTS idx_comentarios_reactions_user_id ON comentarios_reactions (user_id);
CREATE INDEX IF NOT EXISTS idx_comentarios_reactions_tipo ON comentarios_reactions (tipo);

-- =====================================================
-- PASSO 6: Criar VIEWS
-- =====================================================

-- View: Palpites com estatísticas
CREATE OR REPLACE VIEW palpites_stats AS
SELECT 
    p.id,
    p.user_id,
    p.titulo,
    p.img_url,
    p.link,
    p.created_at,
    p.updated_at,
    COALESCE(likes.count, 0) AS total_likes,
    COALESCE(dislikes.count, 0) AS total_dislikes,
    COALESCE(comments.count, 0) AS total_comentarios,
    u.nome AS autor_nome,
    u.avatar AS autor_avatar
FROM palpites p
LEFT JOIN users u ON p.user_id = u.id
LEFT JOIN (
    SELECT palpite_id, COUNT(*) as count 
    FROM palpites_reactions 
    WHERE tipo = 'like' 
    GROUP BY palpite_id
) likes ON p.id = likes.palpite_id
LEFT JOIN (
    SELECT palpite_id, COUNT(*) as count 
    FROM palpites_reactions 
    WHERE tipo = 'dislike' 
    GROUP BY palpite_id
) dislikes ON p.id = dislikes.palpite_id
LEFT JOIN (
    SELECT palpite_id, COUNT(*) as count 
    FROM comentarios 
    GROUP BY palpite_id
) comments ON p.id = comments.palpite_id;

-- View: Comentários com estatísticas
CREATE OR REPLACE VIEW comentarios_stats AS
SELECT 
    c.id,
    c.palpite_id,
    c.user_id,
    c.texto,
    c.created_at,
    c.updated_at,
    COALESCE(likes.count, 0) AS total_likes,
    COALESCE(dislikes.count, 0) AS total_dislikes,
    u.nome AS autor_nome,
    u.avatar AS autor_avatar
FROM comentarios c
LEFT JOIN users u ON c.user_id = u.id
LEFT JOIN (
    SELECT comentario_id, COUNT(*) as count 
    FROM comentarios_reactions 
    WHERE tipo = 'like' 
    GROUP BY comentario_id
) likes ON c.id = likes.comentario_id
LEFT JOIN (
    SELECT comentario_id, COUNT(*) as count 
    FROM comentarios_reactions 
    WHERE tipo = 'dislike' 
    GROUP BY comentario_id
) dislikes ON c.id = dislikes.comentario_id;

-- =====================================================
-- PASSO 7: Criar FUNÇÕES
-- =====================================================

-- Função para toggle de reação em palpite
CREATE OR REPLACE FUNCTION toggle_palpite_reaction(
    p_palpite_id INTEGER,
    p_user_id INTEGER,
    p_tipo VARCHAR(10)
)
RETURNS TABLE (
    action VARCHAR(10),
    total_likes BIGINT,
    total_dislikes BIGINT
) AS $$
DECLARE
    v_existing_tipo VARCHAR(10);
BEGIN
    -- Verificar se já existe uma reação
    SELECT tipo INTO v_existing_tipo
    FROM palpites_reactions
    WHERE palpite_id = p_palpite_id AND user_id = p_user_id;
    
    IF v_existing_tipo IS NULL THEN
        -- Inserir nova reação
        INSERT INTO palpites_reactions (palpite_id, user_id, tipo)
        VALUES (p_palpite_id, p_user_id, p_tipo);
        action := 'added';
    ELSIF v_existing_tipo = p_tipo THEN
        -- Remover reação (toggle off)
        DELETE FROM palpites_reactions
        WHERE palpite_id = p_palpite_id AND user_id = p_user_id;
        action := 'removed';
    ELSE
        -- Alterar tipo de reação
        UPDATE palpites_reactions
        SET tipo = p_tipo
        WHERE palpite_id = p_palpite_id AND user_id = p_user_id;
        action := 'changed';
    END IF;
    
    -- Retornar contadores atualizados
    RETURN QUERY
    SELECT 
        action,
        (SELECT COUNT(*) FROM palpites_reactions WHERE palpite_id = p_palpite_id AND tipo = 'like'),
        (SELECT COUNT(*) FROM palpites_reactions WHERE palpite_id = p_palpite_id AND tipo = 'dislike');
END;
$$ LANGUAGE plpgsql;

-- Função para toggle de reação em comentário
CREATE OR REPLACE FUNCTION toggle_comentario_reaction(
    p_comentario_id INTEGER,
    p_user_id INTEGER,
    p_tipo VARCHAR(10)
)
RETURNS TABLE (
    action VARCHAR(10),
    total_likes BIGINT,
    total_dislikes BIGINT
) AS $$
DECLARE
    v_existing_tipo VARCHAR(10);
BEGIN
    -- Verificar se já existe uma reação
    SELECT tipo INTO v_existing_tipo
    FROM comentarios_reactions
    WHERE comentario_id = p_comentario_id AND user_id = p_user_id;
    
    IF v_existing_tipo IS NULL THEN
        -- Inserir nova reação
        INSERT INTO comentarios_reactions (comentario_id, user_id, tipo)
        VALUES (p_comentario_id, p_user_id, p_tipo);
        action := 'added';
    ELSIF v_existing_tipo = p_tipo THEN
        -- Remover reação (toggle off)
        DELETE FROM comentarios_reactions
        WHERE comentario_id = p_comentario_id AND user_id = p_user_id;
        action := 'removed';
    ELSE
        -- Alterar tipo de reação
        UPDATE comentarios_reactions
        SET tipo = p_tipo
        WHERE comentario_id = p_comentario_id AND user_id = p_user_id;
        action := 'changed';
    END IF;
    
    -- Retornar contadores atualizados
    RETURN QUERY
    SELECT 
        action,
        (SELECT COUNT(*) FROM comentarios_reactions WHERE comentario_id = p_comentario_id AND tipo = 'like'),
        (SELECT COUNT(*) FROM comentarios_reactions WHERE comentario_id = p_comentario_id AND tipo = 'dislike');
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- VERIFICAÇÃO FINAL
-- =====================================================

-- Verificar tabelas criadas
SELECT 
    tablename,
    'Criada com sucesso' as status
FROM pg_tables 
WHERE schemaname = 'public' 
AND tablename IN ('palpites', 'comentarios', 'palpites_reactions', 'comentarios_reactions')
ORDER BY tablename;

-- Verificar views criadas
SELECT 
    viewname,
    'Criada com sucesso' as status
FROM pg_views 
WHERE schemaname = 'public' 
AND viewname IN ('palpites_stats', 'comentarios_stats')
ORDER BY viewname;

-- Verificar funções criadas
SELECT 
    routine_name,
    'Criada com sucesso' as status
FROM information_schema.routines 
WHERE routine_schema = 'public' 
AND routine_name IN ('toggle_palpite_reaction', 'toggle_comentario_reaction')
ORDER BY routine_name;
